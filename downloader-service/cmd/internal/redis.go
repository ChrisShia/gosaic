package internal

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

func saveToRedis(img image.Image, c *redis.Client, ip string, indexPrefix string, indexer func(image.Image) ([3]float64, error), ctx context.Context) error {
	float64Vector, err := indexer(img)
	if err != nil {
		return err
	}
	avColorBinary, err := binaryFloat64bit(float64Vector)
	if err != nil {
		return err
	}

	imgBase64String, err := imageToBase64String(img)
	if err != nil {
		return err
	}

	data := struct {
		Img          string `redis:"img"`
		AverageColor []byte `redis:"average_color"`
	}{
		Img:          imgBase64String,
		AverageColor: avColorBinary,
	}

	counterKey := fmt.Sprintf("%s:counter", ip)

	id, err := c.Incr(ctx, counterKey).Result()
	if err != nil {
		return err
	}

	key := dbKey(indexPrefix, ip, id)

	if err = c.HSet(ctx, key, data).Err(); err != nil {
		return err
	}

	return nil
}

func dbKey(indexPrefix string, ip string, id int64) string {
	return fmt.Sprintf("%s:%s:%d", indexPrefix, ip, id)
}

func binaryFloat64bit(indexVector [3]float64) ([]byte, error) {
	avColorBinary := make([]byte, 8*len(indexVector))
	for i, f := range indexVector {
		binary.LittleEndian.PutUint64(avColorBinary[i*8:], math.Float64bits(f))
	}

	return avColorBinary, nil
}

func imageToBase64String(img image.Image) (string, error) {
	imgBuf := new(bytes.Buffer)
	err := jpeg.Encode(imgBuf, img, nil)
	if err != nil {
		return "", err
	}

	base64Str := base64.StdEncoding.EncodeToString(imgBuf.Bytes())

	return base64Str, nil
}

func RedisFTCREATE(indexName string, c *redis.Client, indexPrefix string) {
	_, err := c.Do(context.Background(),
		"FT.CREATE", indexName,
		"ON", "HASH",
		"PREFIX", "1", indexPrefix,
		"SCHEMA", "average_color", "VECTOR", "HNSW", "6",
		"TYPE", "FLOAT64",
		"DIM", "3",
		"DISTANCE_METRIC", "L2", //Euclidean distance
	).Result()
	if err != nil {
		// ignore error if index already exists
		if err.Error() != "Index already exists" {
			log.Fatalf("FTCreate failed: %s", err)
		}
	}
}

func RedisFTSEARCH(searchFor [3]float64, indexName string, c *redis.Client) (interface{}, error) {
	searchForBinary, err := binaryFloat64bit(searchFor)
	if err != nil {
		return nil, err
	}

	result, err := c.Do(context.Background(),
		"FT.SEARCH", indexName,
		"*=>[KNN 5 @average_color $vec]",
		"PARAMS", "2", "vec", searchForBinary,
		"SORTBY", "average_color",
		"RETURN", "2", "img", "average_color",
		"DIALECT", "2",
	).Result()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func EstablishRedisConnAndPing(addr string) (*redis.Client, error) {
	client, err := RedisClient(addr)
	if err != nil {
		return nil, err
	}

	timeout, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	if err = client.Ping(timeout).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

func RedisClient(addr string) (*redis.Client, error) {
	opt, err := redis.ParseURL(addr)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)
	return client, nil
}

func splitFields(v []byte) <-chan []byte {
	out := make(chan []byte)

	go func() {
		defer close(out)
		remainder := v
		for len(remainder) > 0 {
			keyValue, rem := FirstInLineKeyValue(remainder)
			remainder = rem
			out <- keyValue
		}
	}()

	return out
}

func RedisResults(s []byte) []*Result {
	allResults, _ := redisResultsField(s)

	return Results(allResults)
}

func Results(s []byte) []*Result {
	r := make([]*Result, 0)

	for res := range bridge(resultsStream(s, nil), nil) {
		r = append(r, res)
	}

	return r
}

func bridge(chanStream <-chan <-chan *Result, done <-chan struct{}) <-chan *Result {
	out := make(chan *Result)

	go func() {
		defer close(out)
		for {
			var stream <-chan *Result
			select {
			case v, ok := <-chanStream:
				if !ok {
					return
				}
				stream = v
			case <-done:
				return
			}

			for v := range stream {
				select {
				case out <- v:
				case <-done:
					return
				}
			}
		}
	}()

	return out
}

func resultsStream(s []byte, done <-chan struct{}) <-chan <-chan *Result {
	mapIdentifier := "map"
	f := func(bs []byte) (int, int) {
		return 0, bytes.Index(bs, []byte(mapIdentifier)) + len(mapIdentifier)
	}
	outQueue := queuedStreamsFunc(s, f, done)

	out := make(chan (<-chan *Result))
	go func() {
		defer close(out)
		for _, c := range outQueue {
			out <- c
		}
	}()

	return out
}

func queuedStreamsFunc(s []byte, f func(bs []byte) (int, int), done <-chan struct{}) []<-chan *Result {
	outQueue := make([]<-chan *Result, 0)

	s = s[1 : len(s)-1]

	keyValue := make([]byte, 0)
	for len(s) > 0 {
		if len(s) == 0 {
			break
		}

		keyValue, s = firstAvailableFunc(s, f)

		c := make(chan *Result)
		outQueue = append(outQueue, c)
		go func(kv []byte, o chan<- *Result) {
			defer close(o)
			select {
			case o <- NewResult(kv):
			case <-done:
				return
			}
		}(keyValue, c)
	}

	return outQueue
}

type Result struct {
	Attributes string
	//Attributes map[string]string
}

func NewResult(s []byte) *Result {
	//m := make(map[string]string)

	return &Result{
		Attributes: string(s),
	}
}

func redisResultsField(s []byte) ([]byte, []byte) {
	f := func(bs []byte) (int, int) {
		identifier := []byte("results:")
		i := bytes.Index(s, identifier)
		if i == -1 {
			return -1, -1
		}
		return i, i + len(identifier)
	}

	return firstAvailableFunc(s, f)
}

func FirstInLineKeyValue(s []byte) ([]byte, []byte) {
	f := func(bs []byte) (int, int) {
		return 0, bytes.IndexByte(bs, ':') + 1
	}

	return firstAvailableFunc(s, f)
}

func firstAvailableFunc(s []byte, f func(bs []byte) (int, int)) ([]byte, []byte) {
	if len(s) == 0 {
		return []byte{}, []byte{}
	}

	identifierStart, valueStart := f(s)
	if identifierStart == -1 {
		return []byte{}, []byte{}
	}

	valueEnd := findValueEnd(s, valueStart)
	keyValue := s[identifierStart : valueEnd+1]

	remainder := bytes.TrimLeft(s[valueEnd+1:], " ")
	return keyValue, remainder
}

func findValueEnd(s []byte, start int) int {
	var end int

	switch {
	case isBra(s[start]):
		end = matchingKetIndex(s, start)
	default:
		end = start + indexBeforeFirstWhiteSpaceOrEndOfSequence(s[start:])
	}

	return end
}

func indexBeforeFirstWhiteSpaceOrEndOfSequence(s []byte) int {
	firstWhiteSpace := bytes.IndexRune(s, ' ')
	if firstWhiteSpace == -1 {
		return len(s) - 1
	}

	return firstWhiteSpace - 1
}
