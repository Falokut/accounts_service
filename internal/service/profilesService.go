package service

import (
	"context"

	"github.com/Falokut/grpc_errors"
	profiles_service "github.com/Falokut/profiles_service/pkg/profiles_service/v1/protos"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type profilesService struct {
	service profiles_service.ProfilesServiceV1Client
	conn    *grpc.ClientConn
}

func NewProfilesService(serviceAddr string) (*profilesService, error) {
	cc, err := getProfilesServiceConnection(serviceAddr)
	if err != nil {
		return nil, err
	}
	service := profiles_service.NewProfilesServiceV1Client(cc)
	return &profilesService{
		service: service,
		conn:    cc,
	}, nil
}

func (s *profilesService) Shutdown() error {
	return s.conn.Close()
}

func (s *profilesService) CreateProfile(ctx context.Context, profile Profile) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "profilesService.CreateProfile")
	defer span.Finish()
	var err error
	defer span.SetTag("grpc.status", grpc_errors.GetGrpcCode(err))

	_, err = s.service.CreateProfile(ctx, &profiles_service.CreateProfileRequest{
		AccountID:        profile.AccountID,
		Email:            profile.Email,
		Username:         profile.Username,
		RegistrationDate: timestamppb.New(profile.RegistrationDate),
	})

	return err
}
func (s *profilesService) DeleteProfile(ctx context.Context, accountID string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "profilesService.DeleteProfile")
	defer span.Finish()

	_, err := s.service.DeleteProfile(ctx, &profiles_service.DeleteProfileRequest{AccountID: accountID})
	span.SetTag("grpc.status", status.Code(err))
	return err
}

func getProfilesServiceConnection(serviceAddr string) (*grpc.ClientConn, error) {
	return grpc.Dial(serviceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(
			otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
		grpc.WithStreamInterceptor(
			otgrpc.OpenTracingStreamClientInterceptor(opentracing.GlobalTracer())),
	)
}
