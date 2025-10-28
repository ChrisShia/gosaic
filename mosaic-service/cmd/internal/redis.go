package internal

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"image"
	"math"

	"github.com/redis/go-redis/v9"
)

type RedisIndex struct {
	Name   string
	Prefix string
	Client *redis.Client
}

func NewRedisIndex(name string, prefix string, c *redis.Client) *RedisIndex {
	return &RedisIndex{
		Name:   name,
		Prefix: prefix,
		Client: c,
	}
}

func (ri *RedisIndex) Image(ac [3]float64) (image.Image, error) {
	ftSearchResults, err := ri.FTSEARCH(ac)
	if err != nil {
		return nil, err
	}

	result, err := NearestNeighbourRedisResult(ftSearchResults)
	if err != nil {
		return nil, err
	}

	return base64StringToImage(result)
}

func base64StringToImage(str string) (image.Image, error) {
	var p []byte
	_, err := base64.StdEncoding.Decode(p, []byte(str))
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(p))
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (ri *RedisIndex) FTSEARCH(searchFor [3]float64) (interface{}, error) {
	searchForBinary, err := binaryFloat64bit(searchFor)
	if err != nil {
		return nil, err
	}

	result, err := ri.Client.Do(context.Background(),
		"FT.SEARCH", ri.Name,
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

func (ri *RedisIndex) PipeFTSEARCHAndRemove(searchFor [3]float64) (interface{}, error) {
	//TODO: return nil for now but needs completion...Complete the Pipeline: Retrieval and Removal from redis
	//searchForBinary, err := binaryFloat64bit(searchFor)
	//if err != nil {
	//	return nil, err
	//}
	//
	//ri.Client.TxPipeline().Do(context.Background())
	//
	//ri.Client.TxPipelined(context.Background(), func(pipe redis.Pipeliner) error {
	//
	//	return nil
	//})

	return nil, nil
}

func binaryFloat64bit(indexVector [3]float64) ([]byte, error) {
	avColorBinary := make([]byte, 8*len(indexVector))
	for i, f := range indexVector {
		binary.LittleEndian.PutUint64(avColorBinary[i*8:], math.Float64bits(f))
	}

	return avColorBinary, nil
}

var (
	ErrNoResult          = errors.New("no images found in results")
	ErrInvalidResultType = errors.New("invalid result type")
	ErrNoImageResult     = errors.New("invalid result, no \"img\" field")
	ErrInvalidField      = errors.New("invalid type of \"img\" field")
)

func NearestNeighbourRedisResult(result interface{}) (string, error) {
	redisResultMap := result.(map[interface{}]interface{})

	allResults := redisResultMap["results"].([]interface{})
	if len(allResults) == 0 {
		return "", ErrNoResult
	}

	firstResult := allResults[0]

	var firstResultMap map[interface{}]interface{}
	switch firstResult.(type) {
	case map[interface{}]interface{}:
		firstResultMap = firstResult.(map[interface{}]interface{})
	default:
		return "", ErrInvalidResultType
	}

	firstResultExtraAttributesMap := firstResultMap["extra_attributes"].(map[interface{}]interface{})

	actualImg := firstResultExtraAttributesMap["img"]
	if actualImg == nil {
		return "", ErrNoImageResult
	}

	var res string
	switch actualImg.(type) {
	case string:
		res = actualImg.(string)
	default:
		return "", ErrInvalidField
	}

	return res, nil
}
