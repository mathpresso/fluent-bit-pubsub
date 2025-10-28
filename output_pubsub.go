package main

import (
	"C"
	"fmt"
	"strconv"
	"time"
	"unsafe"

	"cloud.google.com/go/pubsub"

	"context"

	"github.com/fluent/fluent-bit-go/output"

	jsoniter "github.com/json-iterator/go"
)
import "os"

var (
	wrapper = OutputWrapper(&Output{})
)

type Output struct{}

type OutputWrapper interface {
	Register(ctx unsafe.Pointer, name string, desc string) int
	GetConfigKey(ctx unsafe.Pointer, key string) string
	NewDecoder(data unsafe.Pointer, length int) *output.FLBDecoder
	GetRecord(dec *output.FLBDecoder) (ret int, ts interface{}, rec map[interface{}]interface{})
}

func (o *Output) Register(ctx unsafe.Pointer, name string, desc string) int {
	return output.FLBPluginRegister(ctx, name, desc)
}

func (o *Output) GetConfigKey(ctx unsafe.Pointer, key string) string {
	return output.FLBPluginConfigKey(ctx, key)
}

func (o *Output) NewDecoder(data unsafe.Pointer, length int) *output.FLBDecoder {
	return output.NewDecoder(data, length)
}

func (o *Output) GetRecord(dec *output.FLBDecoder) (ret int, ts interface{}, rec map[interface{}]interface{}) {
	return output.GetRecord(dec)
}

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return wrapper.Register(ctx, "pubsub", "output pubsub")
}

//export FLBPluginInit
func FLBPluginInit(ctx unsafe.Pointer) int {
	var err error
	project := wrapper.GetConfigKey(ctx, "Project")
	topic := wrapper.GetConfigKey(ctx, "Topic")
	jwtPath := wrapper.GetConfigKey(ctx, "JwtPath")
	dg := wrapper.GetConfigKey(ctx, "Debug")
	to := wrapper.GetConfigKey(ctx, "Timeout")
	bt := wrapper.GetConfigKey(ctx, "ByteThreshold")
	ct := wrapper.GetConfigKey(ctx, "CountThreshold")
	dt := wrapper.GetConfigKey(ctx, "DelayThreshold")

	fmt.Printf("[pubsub-go] plugin parameter project = '%s'\n", project)
	fmt.Printf("[pubsub-go] plugin parameter topic = '%s'\n", topic)
	fmt.Printf("[pubsub-go] plugin parameter jwtPath = '%s'\n", jwtPath)
	fmt.Printf("[pubsub-go] plugin parameter debug = '%s'\n", dg)
	fmt.Printf("[pubsub-go] plugin parameter timeout = '%s'\n", to)
	fmt.Printf("[pubsub-go] plugin parameter byte threshold = '%s'\n", bt)
	fmt.Printf("[pubsub-go] plugin parameter count threshold = '%s'\n", ct)
	fmt.Printf("[pubsub-go] plugin parameter delay threshold = '%s'\n", dt)

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Printf("[err][init] %+v\n", err)
		return output.FLB_ERROR
	}

	fmt.Printf("[pubsub-go] plugin hostname = '%s'\n", hostname)

	// Parse settings into local variables (not global)
	instanceTimeout := pubsub.DefaultPublishSettings.Timeout
	instanceByteThreshold := pubsub.DefaultPublishSettings.ByteThreshold
	instanceCountThreshold := pubsub.DefaultPublishSettings.CountThreshold
	instanceDelayThreshold := pubsub.DefaultPublishSettings.DelayThreshold

	if dg != "" {
		_, err = strconv.ParseBool(dg)
		if err != nil {
			fmt.Printf("[err][init] %+v\n", err)
			return output.FLB_ERROR
		}
	}
	if to != "" {
		v, err := strconv.Atoi(to)
		if err != nil {
			fmt.Printf("[err][init] %+v\n", err)
			return output.FLB_ERROR
		}
		instanceTimeout = time.Duration(v) * time.Millisecond
	}
	if bt != "" {
		v, err := strconv.Atoi(bt)
		if err != nil {
			fmt.Printf("[err][init] %+v\n", err)
			return output.FLB_ERROR
		}
		instanceByteThreshold = v
	}
	if ct != "" {
		v, err := strconv.Atoi(ct)
		if err != nil {
			fmt.Printf("[err][init] %+v\n", err)
			return output.FLB_ERROR
		}
		instanceCountThreshold = v
	}
	if dt != "" {
		v, err := strconv.Atoi(dt)
		if err != nil {
			fmt.Printf("[err][init] %+v\n", err)
			return output.FLB_ERROR
		}
		instanceDelayThreshold = time.Duration(v) * time.Millisecond
	}
	publishSetting := pubsub.PublishSettings{
		ByteThreshold:  instanceByteThreshold,
		CountThreshold: instanceCountThreshold,
		DelayThreshold: instanceDelayThreshold,
		Timeout:        instanceTimeout,
	}

	keeper, err := NewKeeper(project, topic, jwtPath, &publishSetting)
	if err != nil {
		fmt.Printf("[err][init] %+v\n", err)
		return output.FLB_ERROR
	}

	// Store keeper in Fluent Bit's context
	output.FLBPluginSetContext(ctx, keeper)

	return output.FLB_OK
}

// FLBPluginFlush: 구버전 ABI 엔트리포인트
// 구버전 Fluent Bit에서는 'data'가 플러그인 컨텍스트를 가리킵니다.
// 하위 호환을 위해 최신 방식 함수(아래)로 위임합니다.
//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
    // Call the ctx-aware variant using data as ctx for old ABI compatibility
    return FLBPluginFlushCtx(data, data, length, tag)
}

// FLBPluginFlushCtx: 최신 ABI 엔트리포인트
// 최신 Fluent Bit에서는 플러그인 컨텍스트는 'ctx'로, 레코드 버퍼는 'data'로 전달됩니다.
// 신규 버전에서 컨텍스트 nil 문제가 생기지 않도록 항상 'ctx'에서 컨텍스트를 조회합니다.
//export FLBPluginFlushCtx
func FLBPluginFlushCtx(ctx unsafe.Pointer, data unsafe.Pointer, length C.int, tag *C.char) int {
    ctxData := output.FLBPluginGetContext(ctx)
    if ctxData == nil {
        fmt.Printf("[err][flush] context is nil\n")
        return output.FLB_ERROR
    }
    instanceKeeper, ok := ctxData.(Keeper)
    if !ok {
        fmt.Printf("[err][flush] invalid keeper type\n")
        return output.FLB_ERROR
    }

    bctx := context.Background()
    tagname := ""
    if tag != nil {
        tagname = C.GoString(tag)
    }

    dec := wrapper.NewDecoder(data, int(length))
    var results []*pubsub.PublishResult
    var err error
    var message []byte

    for {
        ret, ts, record := wrapper.GetRecord(dec)
        if ret != 0 {
            break
        }
        timestampStr := fmt.Sprintf("%v", ts)
        record, err = DecodeMap(record)
        if err != nil {
            fmt.Printf("Failed to decode record: [%s] %s %v\n", tagname, timestampStr, record)
        }

        var json = jsoniter.ConfigCompatibleWithStandardLibrary
        message, err = json.Marshal(record)
        if err != nil {
            fmt.Printf("Failed to decode record: [%s] %s %v\n", tagname, timestampStr, record)
        }
        results = append(results, instanceKeeper.Send(bctx, interfaceToBytes(message)))
    }

    for _, result := range results {
        if result != nil {
            if _, err := result.Get(bctx); err != nil {
                if err == context.DeadlineExceeded || err == context.Canceled {
                    fmt.Printf("[err][publish][retry] %+v \n", err)
                    return output.FLB_RETRY
                }
                fmt.Printf("[err][publish][don't retry] %+v \n", err)
            }
        }
    }
    return output.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit(ctx unsafe.Pointer) int {
	ctxData := output.FLBPluginGetContext(ctx)
	if ctxData != nil {
		if instanceKeeper, ok := ctxData.(Keeper); ok {
			instanceKeeper.Stop()
		}
	}
	return output.FLB_OK
}

func interfaceToBytes(v interface{}) []byte {
	switch d := v.(type) {
	case []byte:
		return d
	case string:
		return []byte(d)
	case int, int32, int64, uint, uint32, uint64:
		return []byte(fmt.Sprintf("%d", d))
	case float32, float64:
		return []byte(fmt.Sprintf("%f", d))
	case bool:
		return []byte(strconv.FormatBool(d))
	case time.Time:
		return []byte(d.Format(time.RFC3339))
	default:
		return []byte(fmt.Sprintf("%v", d))
	}
}

// DecodeMap prepares a record for JSON marshalling
// Any []byte will be base64 encoded when marshaled to JSON, so we must directly cast all []byte to string
func DecodeMap(record map[interface{}]interface{}) (map[interface{}]interface{}, error) {
	for k, v := range record {
		switch t := v.(type) {
		case []byte:
			// convert all byte slices to strings
			record[k] = string(t)
		case map[interface{}]interface{}:
			decoded, err := DecodeMap(t)
			if err != nil {
				return nil, err
			}
			record[k] = decoded
		case []interface{}:
			decoded, err := decodeSlice(t)
			if err != nil {
				return nil, err
			}
			record[k] = decoded
		}
	}
	return record, nil
}

func decodeSlice(record []interface{}) ([]interface{}, error) {
	for i, v := range record {
		switch t := v.(type) {
		case []byte:
			// convert all byte slices to strings
			record[i] = string(t)
		case map[interface{}]interface{}:
			decoded, err := DecodeMap(t)
			if err != nil {
				return nil, err
			}
			record[i] = decoded
		case []interface{}:
			decoded, err := decodeSlice(t)
			if err != nil {
				return nil, err
			}
			record[i] = decoded
		}
	}
	return record, nil
}

func main() {}
