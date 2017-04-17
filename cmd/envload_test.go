package cmd

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func TestConvertChunkedParams(t *testing.T) {
	f := func(s int) []*ssm.ParameterMetadata {
		var r []*ssm.ParameterMetadata
		for i := 0; i < s; i += 1 {
			r = append(r, &ssm.ParameterMetadata{Name: aws.String("stg.foo.ENV")})
		}
		return r
	}
	ts := []struct {
		in     []*ssm.ParameterMetadata
		out    int
		chunks []int
	}{
		{f(0), 0, []int{0}},
		{f(1), 1, []int{1}},
		{f(9), 1, []int{9}},
		{f(10), 1, []int{10}},
		{f(11), 2, []int{10, 1}},
	}

	for _, v := range ts {
		res := convertChunkedParams(v.in)
		if len(res) != v.out {
			t.Errorf("expected length %d, got length %v", v.out, len(res))
		}

		for i, vv := range res {
			if len(vv) != v.chunks[i] {
				t.Errorf("expected chunk length %d, got length %v", v.chunks[i], len(vv))
			}
		}
	}
}
