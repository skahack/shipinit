package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/spf13/cobra"
)

type envloadCmd struct {
	env         string
	serviceName string
}

func NewEnvloadCommand(out, errOut io.Writer) *cobra.Command {
	f := &envloadCmd{}
	cmd := &cobra.Command{
		Use:   "envload [options]",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			err := f.execute(cmd, args)
			if err != nil {
				fmt.Fprintf(errOut, "%v\n", err)
			}
		},
	}

	cmd.Flags().StringVar(&f.env, "env", "prd", "ECS Cluster Name")
	cmd.Flags().StringVar(&f.serviceName, "service-name", "", "ECS Service Name")

	return cmd
}

func (c *envloadCmd) execute(_ *cobra.Command, args []string) error {
	if c.serviceName == "" {
		return errors.New("--service-name is required")
	}

	if c.env == "" {
		return errors.New("--env is required")
	}

	sess, err := session.NewSession()
	if err != nil {
		return errors.New(fmt.Sprintf("failed to establish AWS session\n%v", err))
	}

	region := getAWSRegion()
	if region == "" {
		return errors.New("AWS region is not found. please set a AWS_DEFAULT_REGION or AWS_REGION")
	}

	client := ssm.New(sess, &aws.Config{
		Region: aws.String(region),
	})

	keys, err := describeParamKeys(client, c.env, c.serviceName)
	if err != nil {
		return err
	}

	cvParams := convertChunkedParams(keys)

	var params []*ssm.Parameter
	for _, v := range cvParams {
		ps, err := getEnvVars(client, v)
		if err != nil {
			return err
		}
		params = append(params, ps...)
	}

	dumpEnvVars(params)

	return nil
}

func _describeParamKeys(client *ssm.SSM, env string, serviceName string, keys []*ssm.ParameterMetadata, next *string) ([]*ssm.ParameterMetadata, error) {
	filter := &ssm.ParametersFilter{
		Key: aws.String("Name"),
		Values: []*string{
			aws.String(fmt.Sprintf("%s.%s.", env, serviceName)),
		},
	}
	filters := []*ssm.ParametersFilter{filter}

	param := &ssm.DescribeParametersInput{
		Filters:    filters,
		MaxResults: aws.Int64(50),
	}
	if *next != "" {
		param.NextToken = next
	}

	res, err := client.DescribeParameters(param)
	if err != nil {
		return nil, err
	}

	keys = append(keys, res.Parameters...)
	if res.NextToken == nil {
		return keys, nil
	} else {
		return _describeParamKeys(client, env, serviceName, keys, res.NextToken)
	}
}

func describeParamKeys(client *ssm.SSM, env string, serviceName string) ([]*ssm.ParameterMetadata, error) {
	var keys []*ssm.ParameterMetadata

	return _describeParamKeys(client, env, serviceName, keys, aws.String(""))
}

func convertChunkedParams(params []*ssm.ParameterMetadata) [][]*ssm.ParameterMetadata {
	var res [][]*ssm.ParameterMetadata
	if len(params) == 0 {
		return [][]*ssm.ParameterMetadata{}
	}

	size := 10
	chunkSize := (len(params) / (size + 1)) + 1

	for i := 0; i < chunkSize; i += 1 {
		from := i * size
		to := from + size
		if to > len(params) {
			to = len(params)
		}
		res = append(res, params[from:to])
	}

	return res
}

func getEnvVars(client *ssm.SSM, params []*ssm.ParameterMetadata) ([]*ssm.Parameter, error) {
	ps := []*string{}

	for _, v := range params {
		ps = append(ps, v.Name)
	}

	p := &ssm.GetParametersInput{
		Names:          ps,
		WithDecryption: aws.Bool(true),
	}
	res, err := client.GetParameters(p)
	if err != nil {
		return nil, err
	}

	return res.Parameters, nil
}

func dumpEnvVars(params []*ssm.Parameter) {
	for _, v := range params {
		ss := strings.Split(*v.Name, ".")
		fmt.Fprintf(os.Stdout, "export %s=%s\n", ss[len(ss)-1], *v.Value)
	}
}

func getAWSRegion() string {
	if os.Getenv("AWS_REGION") != "" {
		return os.Getenv("AWS_REGION")
	}

	if os.Getenv("AWS_DEFAULT_REGION") != "" {
		return os.Getenv("AWS_DEFAULT_REGION")
	}

	return ""
}
