package cluster //nolint:testpackage // Access internal helpers without exporting them.

import (
"context"
"errors"
"testing"
)

var errStartFailed = errors.New("start failed")

type stubProvisionerForStart struct {
startErr      error
startCalls    int
receivedNames []string
}

func (p *stubProvisionerForStart) Start(_ context.Context, name string) error {
p.startCalls++
p.receivedNames = append(p.receivedNames, name)

return p.startErr
}

func (p *stubProvisionerForStart) Create(context.Context, string) error { return nil }
func (p *stubProvisionerForStart) Delete(context.Context, string) error { return nil }
func (p *stubProvisionerForStart) Stop(context.Context, string) error   { return nil }
func (p *stubProvisionerForStart) List(context.Context) ([]string, error) {
return nil, nil
}

func (p *stubProvisionerForStart) Exists(
context.Context,
string,
) (bool, error) {
return false, nil
}

func (p *stubProvisionerForStart) CallCount() int {
return p.startCalls
}

func TestHandleStartRunE_LoadConfigFailure(t *testing.T) {
testHandlerBadConfigLoad(t, HandleStartRunE)
}

func TestHandleStartRunE_FactoryFailure(t *testing.T) {
testHandlerBadFactory(t, HandleStartRunE)
}

func TestHandleStartRunE_ReturnsErrorWhenProvisionerIsNil(t *testing.T) {
testHandlerNilProvisioner(t, HandleStartRunE)
}

func TestHandleStartRunE_ReturnsErrorWhenClusterNameFails(t *testing.T) {
testHandlerBadClusterName(t, HandleStartRunE, &stubProvisionerForStart{})
}

func TestHandleStartRunE_ReturnsErrorWhenProvisionerStartFails(t *testing.T) {
provisioner := &stubProvisionerForStart{startErr: errStartFailed}
testHandlerOperationFails(t, HandleStartRunE, provisioner, "failed to start cluster", provisioner)
}

func TestHandleStartRunE_Success(t *testing.T) {
provisioner := &stubProvisionerForStart{}
testHandlerSuccess(t, HandleStartRunE, provisioner, "Start cluster...", "cluster started", provisioner)
}

//nolint:paralleltest
func TestNewStartCmd_RunESuccess(t *testing.T) {
provisioner := &stubProvisionerForStart{}
testCmdIntegration(t, NewStartCmd, provisioner, provisioner)
}

//nolint:paralleltest
func TestNewStartCmd_FactoryResolutionError(t *testing.T) {
testCmdFactoryError(t, NewStartCmd)
}
