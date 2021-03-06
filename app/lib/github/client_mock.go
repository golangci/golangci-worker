// Code generated by MockGen. DO NOT EDIT.
// Source: client.go

package github

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	github "github.com/google/go-github/github"
	reflect "reflect"
)

// MockClient is a mock of Client interface
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockClient) EXPECT() *MockClientMockRecorder {
	return _m.recorder
}

// GetPullRequest mocks base method
func (_m *MockClient) GetPullRequest(ctx context.Context, c *Context) (*github.PullRequest, error) {
	ret := _m.ctrl.Call(_m, "GetPullRequest", ctx, c)
	ret0, _ := ret[0].(*github.PullRequest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPullRequest indicates an expected call of GetPullRequest
func (_mr *MockClientMockRecorder) GetPullRequest(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "GetPullRequest", reflect.TypeOf((*MockClient)(nil).GetPullRequest), arg0, arg1)
}

// GetPullRequestComments mocks base method
func (_m *MockClient) GetPullRequestComments(ctx context.Context, c *Context) ([]*github.PullRequestComment, error) {
	ret := _m.ctrl.Call(_m, "GetPullRequestComments", ctx, c)
	ret0, _ := ret[0].([]*github.PullRequestComment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPullRequestComments indicates an expected call of GetPullRequestComments
func (_mr *MockClientMockRecorder) GetPullRequestComments(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "GetPullRequestComments", reflect.TypeOf((*MockClient)(nil).GetPullRequestComments), arg0, arg1)
}

// GetPullRequestPatch mocks base method
func (_m *MockClient) GetPullRequestPatch(ctx context.Context, c *Context) (string, error) {
	ret := _m.ctrl.Call(_m, "GetPullRequestPatch", ctx, c)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPullRequestPatch indicates an expected call of GetPullRequestPatch
func (_mr *MockClientMockRecorder) GetPullRequestPatch(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "GetPullRequestPatch", reflect.TypeOf((*MockClient)(nil).GetPullRequestPatch), arg0, arg1)
}

// CreateReview mocks base method
func (_m *MockClient) CreateReview(ctx context.Context, c *Context, review *github.PullRequestReviewRequest) error {
	ret := _m.ctrl.Call(_m, "CreateReview", ctx, c, review)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateReview indicates an expected call of CreateReview
func (_mr *MockClientMockRecorder) CreateReview(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CreateReview", reflect.TypeOf((*MockClient)(nil).CreateReview), arg0, arg1, arg2)
}

// SetCommitStatus mocks base method
func (_m *MockClient) SetCommitStatus(ctx context.Context, c *Context, ref string, status Status, desc string, url string) error {
	ret := _m.ctrl.Call(_m, "SetCommitStatus", ctx, c, ref, status, desc, url)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetCommitStatus indicates an expected call of SetCommitStatus
func (_mr *MockClientMockRecorder) SetCommitStatus(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SetCommitStatus", reflect.TypeOf((*MockClient)(nil).SetCommitStatus), arg0, arg1, arg2, arg3, arg4, arg5)
}
