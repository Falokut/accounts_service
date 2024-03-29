// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.25.3
// source: accounts_service_v1.proto

package protos

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// AccountsServiceV1Client is the client API for AccountsServiceV1 service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AccountsServiceV1Client interface {
	CreateAccount(ctx context.Context, in *CreateAccountRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	DeleteAccount(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	RequestAccountVerificationToken(ctx context.Context, in *VerificationTokenRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	VerifyAccount(ctx context.Context, in *VerifyAccountRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SignIn(ctx context.Context, in *SignInRequest, opts ...grpc.CallOption) (*AccessResponse, error)
	GetAccountID(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Logout(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	RequestChangePasswordToken(ctx context.Context, in *ChangePasswordTokenRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	ChangePassword(ctx context.Context, in *ChangePasswordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	GetAllSessions(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*AllSessionsResponse, error)
	TerminateSessions(ctx context.Context, in *TerminateSessionsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type accountsServiceV1Client struct {
	cc grpc.ClientConnInterface
}

func NewAccountsServiceV1Client(cc grpc.ClientConnInterface) AccountsServiceV1Client {
	return &accountsServiceV1Client{cc}
}

func (c *accountsServiceV1Client) CreateAccount(ctx context.Context, in *CreateAccountRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/accounts_service.accountsServiceV1/CreateAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountsServiceV1Client) DeleteAccount(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/accounts_service.accountsServiceV1/DeleteAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountsServiceV1Client) RequestAccountVerificationToken(ctx context.Context, in *VerificationTokenRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/accounts_service.accountsServiceV1/RequestAccountVerificationToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountsServiceV1Client) VerifyAccount(ctx context.Context, in *VerifyAccountRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/accounts_service.accountsServiceV1/VerifyAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountsServiceV1Client) SignIn(ctx context.Context, in *SignInRequest, opts ...grpc.CallOption) (*AccessResponse, error) {
	out := new(AccessResponse)
	err := c.cc.Invoke(ctx, "/accounts_service.accountsServiceV1/SignIn", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountsServiceV1Client) GetAccountID(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/accounts_service.accountsServiceV1/GetAccountID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountsServiceV1Client) Logout(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/accounts_service.accountsServiceV1/Logout", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountsServiceV1Client) RequestChangePasswordToken(ctx context.Context, in *ChangePasswordTokenRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/accounts_service.accountsServiceV1/RequestChangePasswordToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountsServiceV1Client) ChangePassword(ctx context.Context, in *ChangePasswordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/accounts_service.accountsServiceV1/ChangePassword", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountsServiceV1Client) GetAllSessions(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*AllSessionsResponse, error) {
	out := new(AllSessionsResponse)
	err := c.cc.Invoke(ctx, "/accounts_service.accountsServiceV1/GetAllSessions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountsServiceV1Client) TerminateSessions(ctx context.Context, in *TerminateSessionsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/accounts_service.accountsServiceV1/TerminateSessions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AccountsServiceV1Server is the server API for AccountsServiceV1 service.
// All implementations must embed UnimplementedAccountsServiceV1Server
// for forward compatibility
type AccountsServiceV1Server interface {
	CreateAccount(context.Context, *CreateAccountRequest) (*emptypb.Empty, error)
	DeleteAccount(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	RequestAccountVerificationToken(context.Context, *VerificationTokenRequest) (*emptypb.Empty, error)
	VerifyAccount(context.Context, *VerifyAccountRequest) (*emptypb.Empty, error)
	SignIn(context.Context, *SignInRequest) (*AccessResponse, error)
	GetAccountID(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	Logout(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	RequestChangePasswordToken(context.Context, *ChangePasswordTokenRequest) (*emptypb.Empty, error)
	ChangePassword(context.Context, *ChangePasswordRequest) (*emptypb.Empty, error)
	GetAllSessions(context.Context, *emptypb.Empty) (*AllSessionsResponse, error)
	TerminateSessions(context.Context, *TerminateSessionsRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedAccountsServiceV1Server()
}

// UnimplementedAccountsServiceV1Server must be embedded to have forward compatible implementations.
type UnimplementedAccountsServiceV1Server struct {
}

func (UnimplementedAccountsServiceV1Server) CreateAccount(context.Context, *CreateAccountRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAccount not implemented")
}
func (UnimplementedAccountsServiceV1Server) DeleteAccount(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteAccount not implemented")
}
func (UnimplementedAccountsServiceV1Server) RequestAccountVerificationToken(context.Context, *VerificationTokenRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestAccountVerificationToken not implemented")
}
func (UnimplementedAccountsServiceV1Server) VerifyAccount(context.Context, *VerifyAccountRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifyAccount not implemented")
}
func (UnimplementedAccountsServiceV1Server) SignIn(context.Context, *SignInRequest) (*AccessResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignIn not implemented")
}
func (UnimplementedAccountsServiceV1Server) GetAccountID(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAccountID not implemented")
}
func (UnimplementedAccountsServiceV1Server) Logout(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Logout not implemented")
}
func (UnimplementedAccountsServiceV1Server) RequestChangePasswordToken(context.Context, *ChangePasswordTokenRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestChangePasswordToken not implemented")
}
func (UnimplementedAccountsServiceV1Server) ChangePassword(context.Context, *ChangePasswordRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangePassword not implemented")
}
func (UnimplementedAccountsServiceV1Server) GetAllSessions(context.Context, *emptypb.Empty) (*AllSessionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllSessions not implemented")
}
func (UnimplementedAccountsServiceV1Server) TerminateSessions(context.Context, *TerminateSessionsRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TerminateSessions not implemented")
}
func (UnimplementedAccountsServiceV1Server) mustEmbedUnimplementedAccountsServiceV1Server() {}

// UnsafeAccountsServiceV1Server may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AccountsServiceV1Server will
// result in compilation errors.
type UnsafeAccountsServiceV1Server interface {
	mustEmbedUnimplementedAccountsServiceV1Server()
}

func RegisterAccountsServiceV1Server(s grpc.ServiceRegistrar, srv AccountsServiceV1Server) {
	s.RegisterService(&AccountsServiceV1_ServiceDesc, srv)
}

func _AccountsServiceV1_CreateAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServiceV1Server).CreateAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/accounts_service.accountsServiceV1/CreateAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServiceV1Server).CreateAccount(ctx, req.(*CreateAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountsServiceV1_DeleteAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServiceV1Server).DeleteAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/accounts_service.accountsServiceV1/DeleteAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServiceV1Server).DeleteAccount(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountsServiceV1_RequestAccountVerificationToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerificationTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServiceV1Server).RequestAccountVerificationToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/accounts_service.accountsServiceV1/RequestAccountVerificationToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServiceV1Server).RequestAccountVerificationToken(ctx, req.(*VerificationTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountsServiceV1_VerifyAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerifyAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServiceV1Server).VerifyAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/accounts_service.accountsServiceV1/VerifyAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServiceV1Server).VerifyAccount(ctx, req.(*VerifyAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountsServiceV1_SignIn_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignInRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServiceV1Server).SignIn(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/accounts_service.accountsServiceV1/SignIn",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServiceV1Server).SignIn(ctx, req.(*SignInRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountsServiceV1_GetAccountID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServiceV1Server).GetAccountID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/accounts_service.accountsServiceV1/GetAccountID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServiceV1Server).GetAccountID(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountsServiceV1_Logout_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServiceV1Server).Logout(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/accounts_service.accountsServiceV1/Logout",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServiceV1Server).Logout(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountsServiceV1_RequestChangePasswordToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChangePasswordTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServiceV1Server).RequestChangePasswordToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/accounts_service.accountsServiceV1/RequestChangePasswordToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServiceV1Server).RequestChangePasswordToken(ctx, req.(*ChangePasswordTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountsServiceV1_ChangePassword_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChangePasswordRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServiceV1Server).ChangePassword(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/accounts_service.accountsServiceV1/ChangePassword",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServiceV1Server).ChangePassword(ctx, req.(*ChangePasswordRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountsServiceV1_GetAllSessions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServiceV1Server).GetAllSessions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/accounts_service.accountsServiceV1/GetAllSessions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServiceV1Server).GetAllSessions(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountsServiceV1_TerminateSessions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TerminateSessionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServiceV1Server).TerminateSessions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/accounts_service.accountsServiceV1/TerminateSessions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServiceV1Server).TerminateSessions(ctx, req.(*TerminateSessionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AccountsServiceV1_ServiceDesc is the grpc.ServiceDesc for AccountsServiceV1 service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AccountsServiceV1_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "accounts_service.accountsServiceV1",
	HandlerType: (*AccountsServiceV1Server)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateAccount",
			Handler:    _AccountsServiceV1_CreateAccount_Handler,
		},
		{
			MethodName: "DeleteAccount",
			Handler:    _AccountsServiceV1_DeleteAccount_Handler,
		},
		{
			MethodName: "RequestAccountVerificationToken",
			Handler:    _AccountsServiceV1_RequestAccountVerificationToken_Handler,
		},
		{
			MethodName: "VerifyAccount",
			Handler:    _AccountsServiceV1_VerifyAccount_Handler,
		},
		{
			MethodName: "SignIn",
			Handler:    _AccountsServiceV1_SignIn_Handler,
		},
		{
			MethodName: "GetAccountID",
			Handler:    _AccountsServiceV1_GetAccountID_Handler,
		},
		{
			MethodName: "Logout",
			Handler:    _AccountsServiceV1_Logout_Handler,
		},
		{
			MethodName: "RequestChangePasswordToken",
			Handler:    _AccountsServiceV1_RequestChangePasswordToken_Handler,
		},
		{
			MethodName: "ChangePassword",
			Handler:    _AccountsServiceV1_ChangePassword_Handler,
		},
		{
			MethodName: "GetAllSessions",
			Handler:    _AccountsServiceV1_GetAllSessions_Handler,
		},
		{
			MethodName: "TerminateSessions",
			Handler:    _AccountsServiceV1_TerminateSessions_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "accounts_service_v1.proto",
}
