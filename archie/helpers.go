package archie

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog"
	"math"
	"strings"
)

func parseEventPath(key string) (string, string) {
	eventPath := strings.SplitN(key, "/", 2)
	return eventPath[0], eventPath[1]
}

func isJSON(msg []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(msg, &js) == nil
}

func size(s int64) string {
	baseSizes := []map[int64]string{
		{12: "TB"},
		{9: "GB"},
		{6: "MB"},
		{3: "KB"},
		{0: "B"},
	}

	for _, bSize := range baseSizes {
		for exp, unit := range bSize {
			totalSize := float64(s) / math.Pow(10, float64(exp))
			if totalSize >= 1.0 {
				return fmt.Sprintf("%.2f%s", totalSize, unit)
			}
		}
	}

	return "-"
}

func rate(size int64, seconds float64) string {
	bytesPerSecond := float64(size) / seconds

	baseRates := []map[int64]string{
		{12: "TB/s"},
		{9: "GB/s"},
		{6: "MB/s"},
		{3: "KB/s"},
		{0: "B/s"},
	}

	for _, bRate := range baseRates {
		for exp, unit := range bRate {
			totalRate := bytesPerSecond / math.Pow(10, float64(exp))
			if totalRate >= 1.0 {
				return fmt.Sprintf("%.2f%s", totalRate, unit)
			}
		}
	}

	return "-"
}

func logS3Error(err error, msg string, mLog *zerolog.Logger) (error string, code string) {
	s3Err := minio.ToErrorResponse(err)

	if (s3Err == minio.ErrorResponse{}) {
		mLog.Error().Err(err).Msg(msg)
		return err.Error(), msg
	} else {
		if s3Err.Resource == "" {
			mLog.Error().Err(s3Err).
				Dict("s3Error", zerolog.Dict().
					Str("code", s3Err.Code).
					Str("error", s3Err.Message).
					Str("requestID", s3Err.RequestID),
				).
				Msg(msg)
		} else {
			mLog.Error().Err(s3Err).
				Dict("s3Error", zerolog.Dict().
					Str("code", s3Err.Code).
					Str("error", s3Err.Message).
					Str("requestID", s3Err.RequestID).
					Str("resource", s3Err.Resource),
				).
				Msg(msg)
		}
		return s3Err.Message, s3Err.Code
	}
}

func checkContextDone(ctx context.Context) bool {
	// non-blocking
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
