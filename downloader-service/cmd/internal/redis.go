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

	counterKey := ip + ":counter"
	//counterKey := fmt.Sprintf("%s:counter", ip)

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

	return base64.StdEncoding.EncodeToString(imgBuf.Bytes()), nil
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

var mapIdentifier = []byte("map[")

func stripFirstMapIdentifier(results []byte) []byte {
	firstInstanceIndex := bytes.Index(results, mapIdentifier)
	if firstInstanceIndex != 0 {
		return nil
	}

	return results[len(mapIdentifier) : len(results)-1]
}

func splitFields(v []byte) <-chan []byte {
	out := make(chan []byte)

	go func() {
		defer close(out)
		remainder := v
		for len(remainder) > 0 {
			keyValue, rem := firstAvailable(remainder)
			remainder = rem
			out <- keyValue
		}
	}()

	return out
}

func firstAvailable(s []byte) ([]byte, []byte) {
	if len(s) == 0 {
		return []byte{}, []byte{}
	}

	i := bytes.IndexRune(s, ':')
	if i == -1 {
		return []byte{}, []byte{}
	}

	valueStart := i + 1
	valueEnd := findValueEnd(s, valueStart)

	keyValue := s[:valueEnd+1]

	s = bytes.TrimLeft(s[valueEnd+1:], " ")
	return keyValue, s
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

type RR struct {
	attributes   []string
	format       string
	results      []map[string]any
	totalResults string
	warning      string
}

type keyValueString struct {
	key   string
	value string
}

type keyValueMap struct {
	key   string
	value string
}

type result struct {
	extraAttributes map[string]any
}
