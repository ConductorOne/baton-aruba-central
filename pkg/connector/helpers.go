package connector

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/conductorone/baton-aruba-central/pkg/arubacentral"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"google.golang.org/protobuf/types/known/structpb"
)

const ResourcesPageSize uint = 50

func annotationsForUserResourceType() annotations.Annotations {
	annos := annotations.Annotations{}
	annos.Update(&v2.SkipEntitlementsAndGrants{})
	return annos
}

func parsePageToken(i string, resourceID *v2.ResourceId) (*pagination.Bag, uint, error) {
	b := &pagination.Bag{}
	err := b.Unmarshal(i)
	if err != nil {
		return nil, 0, err
	}

	if b.Current() == nil {
		b.Push(pagination.PageState{
			ResourceTypeID: resourceID.ResourceType,
			ResourceID:     resourceID.Resource,
		})
	}

	page, err := convertPageToken(b.PageToken())
	if err != nil {
		return nil, 0, err
	}

	return b, page, nil
}

// convertPageToken converts a string token into an uint.
func convertPageToken(token string) (uint, error) {
	if token == "" {
		return 0, nil
	}

	page, err := strconv.ParseUint(token, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse page token: %w", err)
	}

	return uint(page), nil
}

// prepareNextToken prepares the next page token.
// List responses return number of total items across all pages.
// Offset is zero based index of items, not pages.
// This means we have to increase the offset by the number of items per page to get to the next page.
func prepareNextToken(offset, total uint) string {
	var token string

	if total == 0 {
		return token
	}

	newOffset := offset + ResourcesPageSize
	if newOffset >= total {
		return token
	}

	return fmt.Sprint(newOffset)
}

func getValuesFromProfile[T arubacentral.ProtoValue](profile *structpb.Struct, k string, val T) ([]T, bool) {
	var values []T
	if profile == nil {
		return nil, false
	}

	v, ok := profile.Fields[k]
	if !ok {
		return nil, false
	}

	s, ok := v.Kind.(*structpb.Value_ListValue)
	if !ok {
		return nil, false
	}

	for _, v := range s.ListValue.Values {
		err := val.Unmarshall(v)
		if err != nil {
			return nil, false
		}

		values = append(values, val)
	}

	return values, true
}

func slugify(s string) string {
	lower := strings.ToLower(s)
	return strings.ReplaceAll(lower, " ", "-")
}
