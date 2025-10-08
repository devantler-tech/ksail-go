package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
"context"
"errors"
"testing"
)

var errCreateFailed = errors.New("create failed")

type stubProvisioner struct {
createErr     error
createCalls   int
receivedNames []string
}

func (p *stubProvisioner) Create(_ context.Context, name string) error {
p.createCalls++
p.receivedNames = append(p.receivedNames, name)

return p.createErr
}

func (p *stubProvisioner) Delete(context.Context, string) error { return nil }
func (p *stubProvisioner) Start(context.Context, string) error  { return nil }
func (p *stubProvisioner) Stop(context.Context, string) error   { return nil }
func (p *stubProvisioner) List(context.Context) ([]string, error) {
return nil, nil
}

func (p *stubProvisioner) Exists(context.Context, string) (bool, error) {
return false, nil
}

func (p *stubProvisioner) CallCount() int {
return p.createCalls
}

func TestHandleCreateRunE_LoadConfigFailure(t *testing.T) {
testHandlerBadConfigLoad(t, HandleCreateRunE)
}

func TestHandleCreateRunE_FactoryFailure(t *testing.T) {
testHandlerBadFactory(t, HandleCreateRunE)
}

func TestHandleCreateRunE_ReturnsErrorWhenProvisionerIsNil(t *testing.T) {
testHandlerNilProvisioner(t, HandleCreateRunE)
}

func TestHandleCreateRunE_ReturnsErrorWhenClusterNameFails(t *testing.T) {
testHandlerBadClusterName(t, HandleCreateRunE, &stubProvisioner{})
}

func TestHandleCreateRunE_ReturnsErrorWhenProvisionerCreateFails(t *testing.T) {
provisioner := &stubProvisioner{createErr: errCreateFailed}
testHandlerOperationFails(t, HandleCreateRunE, provisioner, "failed to create cluster", provisioner)
}

func TestHandleCreateRunE_Success(t *testing.T) {
provisioner := &stubProvisioner{}
testHandlerSuccess(t, HandleCreateRunE, provisioner, "Create cluster...", "cluster created", provisioner)
}

//nolint:paralleltest
func TestNewCreateCmd_RunESuccess(t *testing.T) {
provisioner := &stubProvisioner{}
testCmdIntegration(t, NewCreateCmd, provisioner, provisioner)
}
