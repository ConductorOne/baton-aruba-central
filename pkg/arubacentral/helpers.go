package arubacentral

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// toStructpbValue converts a native Go type to a structpb.Value.
func toStructpbValue(val interface{}) (*structpb.Value, error) {
	switch v := val.(type) {
	case string:
		return structpb.NewStringValue(v), nil

	case int:
		return structpb.NewNumberValue(float64(v)), nil

	case float64:
		return structpb.NewNumberValue(v), nil

	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}

func extractRateLimitData(statusCode int, header *http.Header) (*v2.RateLimitDescription, error) {
	if header == nil {
		return nil, nil
	}

	var rlstatus v2.RateLimitDescription_Status

	var limitSecond int64
	var err error
	limitSecondStr := header.Get("X-Ratelimit-Limit-second")
	if limitSecondStr != "" {
		limitSecond, err = strconv.ParseInt(limitSecondStr, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	var limitDay int64
	limitDayStr := header.Get("X-Ratelimit-Limit-day")
	if limitDayStr != "" {
		limitDay, err = strconv.ParseInt(limitDayStr, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	var remainingInSecond int64
	remainingInSecondStr := header.Get("X-Ratelimit-Remaining-second")
	if remainingInSecondStr != "" {
		remainingInSecond, err = strconv.ParseInt(remainingInSecondStr, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	var remainingInDay int64
	remainingInDayStr := header.Get("X-Ratelimit-Remaining-day")
	if remainingInDayStr != "" {
		remainingInDay, err = strconv.ParseInt(remainingInDayStr, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	var resetAt *timestamppb.Timestamp
	var limit, remaining int64
	switch {
	case remainingInSecond == 0:
		rlstatus = v2.RateLimitDescription_STATUS_OVERLIMIT
		limit = limitSecond
		remaining = remainingInSecond
		// if we pass the limit of the second, reset at the next second
		resetAt = timestamppb.New(time.Now().Add(time.Second))
	case remainingInDay == 0:
		rlstatus = v2.RateLimitDescription_STATUS_OVERLIMIT
		limit = limitDay
		remaining = remainingInDay
		now, err := time.Parse(time.RFC1123, header.Get("Date"))
		if err != nil {
			return nil, err
		}
		// if we pass the limit of the day, reset at the next day
		resetAt = timestamppb.New(time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC))
	default:
		rlstatus = v2.RateLimitDescription_STATUS_OK
		limit = limitDay
		remaining = remainingInDay
		now, err := time.Parse(time.RFC1123, header.Get("Date"))
		if err != nil {
			return nil, err
		}
		resetAt = timestamppb.New(time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC))
	}

	return &v2.RateLimitDescription{
		Status:    rlstatus,
		Limit:     limit,
		Remaining: remaining,
		ResetAt:   resetAt,
	}, nil
}
