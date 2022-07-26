package flagstest

import (
	"fmt"
	"testing"

	"github.com/rancher/rancher/tests/framework/clients/rancher"
	"github.com/rancher/rancher/tests/framework/pkg/environmentflag"
	"github.com/rancher/rancher/tests/framework/pkg/session"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type FlagTestSuite struct {
	suite.Suite
	client  *rancher.Client
	session *session.Session
}

func (f *FlagTestSuite) TearDownSuite() {
	f.session.Cleanup()
}

func (f *FlagTestSuite) SetupSuite() {
	testSession := session.NewSession(f.T())
	f.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(f.T(), err)

	f.client = client
}

func (f *FlagTestSuite) TestIstioChart() {
	subSession := f.session.NewSession()
	defer subSession.Cleanup()

	client, err := f.client.WithSession(subSession)
	require.NoError(f.T(), err)

	fmt.Printf("This is Flags Map Value in a fake test: %+v\n", client.Flags)

	if client.Flags.GetValue(environmentflag.Ingress) {
		fmt.Printf("\nIngress flag is enabled in the test\n")
	} else {
		fmt.Printf("\nIngress flag is disabled in the test\n")
	}

	if client.Flags.GetValue(environmentflag.Chart) {
		fmt.Printf("\nChart flag is enabled in the test\n")
	} else {
		fmt.Printf("\nChart flag is disabled in the test\n")
	}
}

func TestFlagTestSuite(t *testing.T) {
	suite.Run(t, new(FlagTestSuite))
}
