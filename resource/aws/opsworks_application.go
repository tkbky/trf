package awsresource

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/aws/aws-sdk-go/service/opsworks/opsworksiface"
	"github.com/tkbky/trf/tf"
)

// OpsWorksApplication describes an application
type OpsWorksApplication struct {
	ID                  *string           `json:"id"`
	Name                *string           `json:"name"`
	ShortName           *string           `json:"short_name"`
	StackID             *string           `json:"stack_id"`
	AppType             *string           `json:"type"`
	Description         *string           `json:"description"`
	Domains             []*string         `json:"domains"`
	Environment         []*Environment    `json:"environment"`
	AppSource           *AppSource        `json:"app_source"`
	DataSource          []*DataSource     `json:"data_source"`
	EnableSSL           *bool             `json:"enable_ssl"`
	SSLConfiguration    *SSLConfiguration `json:"ssl_configuration"`
	DocumentRoot        *string           `json:"document_root"`
	AutoBundleOnDeploy  *string           `json:"auto_bundle_on_deploy"`
	RailsEnv            *string           `json:"rails_env"`
	AwsFlowRubySettings *bool             `json:"aws_flow_ruby_settings"`
}

// Environment describes an applicaiton environment
type Environment struct {
	Key    *string `json:"key"`
	Value  *string `json:"value"`
	Secure *bool   `json:"secure"`
}

// AppSource describes an application source
type AppSource struct {
	SourceType *string `json:"type"`
	URL        *string `json:"url"`
	Username   *string `json:"username"`
	Password   *string `json:"password"`
	SSHKey     *string `json:"ssh_key"`
	Revision   *string `json:"revision"`
}

// SSLConfiguration describes an SSL configuration
type SSLConfiguration struct {
	PrivateKey  *string `json:"private_key"`
	Certificate *string `json:"certificate"`
	Chain       *string `json:"chain"`
}

// DataSource describes a data source
type DataSource struct {
	DataSourceArn          *string `json:"data_source_arn"`
	DataSourceType         *string `json:"data_source_type"`
	DataSourceDatabaseName *string `json:"data_source_database_name"`
}

func newOpsWorksApplication(a *opsworks.App) OpsWorksApplication {
	app := OpsWorksApplication{
		ID:          a.AppId,
		Name:        a.Name,
		ShortName:   a.Shortname,
		StackID:     a.StackId,
		AppType:     a.Type,
		Description: a.Description,
		Domains:     a.Domains,
		AppSource: &AppSource{
			SourceType: a.AppSource.Type,
			URL:        a.AppSource.Url,
			Username:   a.AppSource.Username,
			Password:   a.AppSource.Password,
			SSHKey:     a.AppSource.SshKey,
			Revision:   a.AppSource.Revision,
		},
		EnableSSL: a.EnableSsl,
		SSLConfiguration: &SSLConfiguration{
			PrivateKey:  a.SslConfiguration.PrivateKey,
			Certificate: a.SslConfiguration.Certificate,
			Chain:       a.SslConfiguration.Chain,
		},
		DocumentRoot:       a.Attributes["DocumentRoot"],
		AutoBundleOnDeploy: a.Attributes["AutoBundleOnDeploy"],
		RailsEnv:           a.Attributes["RailsEnv"],
	}

	for _, e := range a.Environment {
		app.Environment = append(app.Environment, &Environment{
			Key:    e.Key,
			Value:  e.Value,
			Secure: e.Secure,
		})
	}

	for _, d := range a.DataSources {
		app.DataSource = append(app.DataSource, &DataSource{
			DataSourceArn:          d.Arn,
			DataSourceType:         d.Type,
			DataSourceDatabaseName: d.DatabaseName,
		})
	}

	return app
}

// DescribeOpsWorksApplication returns a tf string that describes opsworks custom layer
func DescribeOpsWorksApplication(svc opsworksiface.OpsWorksAPI) ([]string, error) {
	resp, err := svc.DescribeStacks(nil)

	if err != nil {
		fmt.Println("Fail to describe application", err)
		return nil, err
	}

	var apps []OpsWorksApplication
	for _, s := range resp.Stacks {
		resp, err := svc.DescribeApps(&opsworks.DescribeAppsInput{StackId: aws.String(*s.StackId)})

		if err != nil {
			fmt.Printf("Fail to describe layer for stack `%s`: %s", *s.StackId, err)
			return nil, err
		}

		for _, a := range resp.Apps {
			apps = append(apps, newOpsWorksApplication(a))
		}
	}

	output := []string{}

	for _, a := range apps {
		resource := tf.Resource{Kind: "aws_opsworks_application", Name: *a.Name, Obj: a}
		output = append(output, fmt.Sprintf("%s", resource.Tf()))
	}

	return output, nil
}
