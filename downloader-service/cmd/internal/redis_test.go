package internal

import (
	"context"
	"encoding/binary"
	"fmt"
	"image"
	"log"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/redis/go-redis/v9"
)

const testImage200300 = "test/test_image_200300.jpg"

const testImagesDir = "test"

const redisTestURL = "redis://localhost:6378"

func Test_RedisHGetAll(t *testing.T) {
	redisClient, closer := redisTestClient()
	defer closer()
	indexPrefix := "img"
	ip := "0.0.0.0"

	RedisFTCREATE("average_color_index", redisClient, indexPrefix)

	testImg := testImage()

	expectedAverageColorVector, err := averageColor(testImg)
	if err != nil {
		t.Errorf("Error calculating average color: %v", err)
	}

	err = saveToRedis(testImg, redisClient, ip, indexPrefix, averageColor, context.Background())
	if err != nil {
		t.Error(err)
	}

	result, err := redisClient.HGetAll(context.Background(), dbKey(indexPrefix, ip, 1)).Result()
	if err != nil {
		t.Error(err)
	}

	averageBinary := result["average_color"]

	redisAverageColorVector, err := float64Vector([]byte(averageBinary))
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expectedAverageColorVector, redisAverageColorVector) {
		t.Errorf("expected average %v, got %v", expectedAverageColorVector, redisAverageColorVector)
	}
}

func Test_RedisDoFTSearch(t *testing.T) {
	redisClient, closer := redisTestClient()
	defer closer()
	indexPrefix := "img"
	ip := "0.0.0.0"

	indexName := "average_color_index"

	RedisFTCREATE(indexName, redisClient, indexPrefix)

	testImg := testImage()

	expectedAverageColorVector, err := averageColor(testImg)
	if err != nil {
		t.Errorf("Error calculating average color: %v", err)
	}

	expectedBinaryRepresentation, err := binaryFloat64bit(expectedAverageColorVector)
	if err != nil {
		t.Errorf("Error calculating binary representation of expected average color: %v", err)
	}

	err = saveToRedis(testImg, redisClient, ip, indexPrefix, averageColor, context.Background())
	if err != nil {
		t.Error(err)
	}

	result, err := redisClient.Do(context.Background(),
		"FT.SEARCH", "average_color_index",
		"*=>[KNN 5 @average_color $vec]",
		"PARAMS", "2", "vec", expectedBinaryRepresentation,
		"SORTBY", "average_color",
		"RETURN", "2", "img", "average_color",
		"DIALECT", "2",
	).Result()
	if err != nil {
		t.Error(err)
	}

	//results := result.(map[interface{}]interface{})

	fmt.Println(result)
}

// Test_RedisFTSearch: Not runnable. Unstable redis command
func Test_RedisFTSearch(t *testing.T) {
	redisClient, closer := redisTestClient()
	defer closer()
	indexPrefix := "img"
	ip := "0.0.0.0"

	indexName := "average_color_index"

	RedisFTCREATE(indexName, redisClient, indexPrefix)

	testImg := testImage()

	expectedAverageColorVector, err := averageColor(testImg)
	if err != nil {
		t.Errorf("Error calculating average color: %v", err)
	}

	expectedVectorBinaryRepresentation, err := binaryFloat64bit(expectedAverageColorVector)
	if err != nil {
		t.Errorf("Error calculating binary representation of expected average color: %v", err)
	}

	err = saveToRedis(testImg, redisClient, ip, indexPrefix, averageColor, context.Background())
	if err != nil {
		t.Error(err)
	}

	//NOTE: Unstable command, should set the flag UnstableResp3 for it to work.
	FTSearch(expectedVectorBinaryRepresentation, redisClient, indexName, context.Background())
	//if err != nil {
	//	t.Error(err)
	//}
	//
	//results := result.(map[interface{}]interface{})
	//
	//fmt.Println(results)
}

func FTSearch(search []byte, c *redis.Client, indexName string, ctx context.Context) {
	result, err := c.FTSearchWithArgs(ctx,
		indexName,
		"*=>[KNN 3 @embedding $vec]",
		&redis.FTSearchOptions{
			Return: []redis.FTSearchReturn{
				{FieldName: "average_color"},
			},
			DialectVersion: 2,
			Params: map[string]any{
				"vec": search,
			},
		},
	).Result()
	if err != nil {
		log.Fatalf("Failed index search: %v", err)
		return
	}

	for _, doc := range result.Docs {
		fmt.Println(doc.Fields)
	}
}

func float64Vector(avBinaryVector []byte) ([3]float64, error) {
	avColor := [3]float64{}
	for i, _ := range avColor {
		start := i * 8
		end := start + 8
		float64bits := avBinaryVector[start:end]
		uInt64 := binary.LittleEndian.Uint64(float64bits)
		float64Frombits := math.Float64frombits(uInt64)
		avColor[i] = float64Frombits
	}

	return avColor, nil
}

func redisTestClient() (*redis.Client, func()) {
	client, err := EstablishRedisConnAndPing(redisTestURL)
	if err != nil {
		log.Fatal(err)
	}

	return client, func() {
		client.FlushAll(context.Background())
		if err := client.Close(); err != nil {
			log.Printf("failed to terminate redis: %s", err)
		}
	}
}

func averageColor(img image.Image) ([3]float64, error) {
	bounds := img.Bounds()
	r, g, b := 0.0, 0.0, 0.0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r1, g1, b1, _ := img.At(x, y).RGBA()
			r, g, b = r+float64(r1), g+float64(g1), b+float64(b1)
		}
	}
	totalPixels := float64(bounds.Max.X * bounds.Max.Y)

	return [3]float64{r / totalPixels, g / totalPixels, b / totalPixels}, nil
}

func testImage() image.Image {
	file, err := os.Open(testImage200300)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	return img
}

func Test(t *testing.T) {
	//[
	//	attributes:[]
	//	format:STRING
	//	results:map[
	//				extra_attributes:map[
	//									average_color:Y�c�Pe�@�^)#@�@m{�Z.��@
	//									img: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
	//									]
	//				id:img:0.0.0.0:1
	//				values:[]
	//			]
	//	total_results:1
	//	warning:[]
	//]
	fmt.Println(map[string]any{"A": "B", "C": "D", "E": 2, "F": map[int]string{1: "A", 2: "B", 3: "C"}})
}

func Test_SplitFields(t *testing.T) {
	var tt = []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{"",
			"map[attributes:[] format:STRING results:[map[extra_attributes:map[average_color:Y�c�Pe�@\\u001B�^)#@�@m{�Z.��@ img:/9j/2wCEAAgGBgcGBQgHBw] id:img:0.0.0.0:1 values:[]]] total_results:1 warning:[]]",
			map[string]string{},
		},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {

		})
	}
}

func Test_splitFields(t *testing.T) {
	var tt = []struct {
		name     string
		input    string
		expected []string
	}{
		{"",
			"attributes:[] format:STRING results:[map[extra_attributes:map[average_color:Y�c�Pe�@\u001B�^)#@�@m{�Z.��@ img:/9j/2wCEAAgGBgcGBQgHBw] id:img:0.0.0.0:1 values:[]]] total_results:1 warning:[]",
			[]string{
				"attributes:[]",
				"format:STRING",
				"results:[map[extra_attributes:map[average_color:Y�c�Pe�@\u001B�^)#@�@m{�Z.��@ img:/9j/2wCEAAgGBgcGBQgHBw] id:img:0.0.0.0:1 values:[]]]",
				"total_results:1",
				"warning:[]",
			},
		},
	}

	for _, tc := range tt {
		ss := make([]map[string]string, 0)
		ss = append(ss, map[string]string{"A": "B"})
		ss = append(ss, map[string]string{"C": "D"})
		fmt.Println(ss)
		t.Run(tc.name, func(t *testing.T) {
			results := make([]string, 0)
			for f := range splitFields([]byte(tc.input)) {
				results = append(results, string(f))
			}

			if len(results) != len(tc.expected) {
				t.Errorf("wrong number of results, expected %d, got %d", len(tc.expected), len(results))
			}

			for i, e := range tc.expected {
				if e != results[i] {
					t.Errorf("got: %s, expected: %s", results[i], e)
				}
			}
		})
	}
}

func Test_FindValueEnd(t *testing.T) {
	var tt = []struct {
		name                  string
		inputString           string
		valueStartIndex       int
		expectedValueEndIndex int
	}{
		{"", "key1:value1 key2:value2", 5, 10},
		{"", "key1:value1 key2:value2", 17, 22},
		{"", "key1:value1 key2:[value2a, value2b]", 17, 34},
		{"", "key1:value1", 5, 10},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			end := findValueEnd([]byte(tc.inputString), tc.valueStartIndex)
			if end != tc.expectedValueEndIndex {
				t.Errorf("got %d, want %d", end, tc.expectedValueEndIndex)
			}
		})
	}
}

func Test_FirstAvailableKeyValuePair(t *testing.T) {
	var tt = []struct {
		name                 string
		input                string
		expectedKeyValuePair string
		expectedRemainder    string
	}{
		{"", "key1:value1 key2:value2", "key1:value1", "key2:value2"},
		{"", "key1:value1 ", "key1:value1", ""},
		{"", "key1:[value1, value1a, value1b] key2:value2", "key1:[value1, value1a, value1b]", "key2:value2"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			pair, remainder := firstAvailable([]byte(tc.input))
			if len(pair) != len(tc.expectedKeyValuePair) {
				t.Errorf("got KeyValue %s, want %s", pair, tc.expectedKeyValuePair)
			} else if len(pair) > 0 {
				for i := 0; i < len(pair); i++ {
					if pair[i] != tc.expectedKeyValuePair[i] {
						t.Errorf("difference in KeyValue at %d: got %s, want %s", i, pair, tc.expectedKeyValuePair)
					}
				}
			}

			if len(remainder) != len(tc.expectedRemainder) {
				t.Errorf("got Remainder %s, want %s", remainder, tc.expectedRemainder)
			} else if len(remainder) > 0 {
				for i := 0; i < len(remainder); i++ {
					if remainder[i] != tc.expectedRemainder[i] {
						t.Errorf("difference in Remainder at %d: got %s, want %s", i, remainder, tc.expectedRemainder)
					}
				}
			}
		})
	}
}
