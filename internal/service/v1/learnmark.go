package v1

import (
	"context"

	"github.com/google/uuid"
	generalv1 "github.com/learnmark/learnmark/api/general/v1"
	learnmarkv1 "github.com/learnmark/learnmark/api/learnmark/v1"
	userv1 "github.com/learnmark/learnmark/api/user/v1"
	"github.com/learnmark/learnmark/internal/dao"
	"github.com/learnmark/learnmark/pkg/build"
	"github.com/learnmark/learnmark/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type learnmarkService struct {
	learnmarkv1.UnimplementedLearnmarkServer
	userDao dao.UserDao
}

func NewlearnmarkService(dao dao.Interface) *learnmarkService {
	return &learnmarkService{userDao: dao.UserDao()}
}

func (s *learnmarkService) GetVersion(context.Context, *emptypb.Empty) (*generalv1.VersionRes, error) {
	ver := build.Version()
	return &generalv1.VersionRes{
		Version: &generalv1.Version{
			Version:   ver.Version,
			GitCommit: ver.GitCommit,
			BuildDate: ver.BuildDate,
			GoVersion: ver.GoVersion,
		},
	}, nil
}

func (s *learnmarkService) SignIn(_ context.Context, req *userv1.SignInReq) (*userv1.SignInRes, error) {
	user, err := s.userDao.SignIn(req.Name, req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to sign in")
	}
	if user.Id == uuid.Nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	accessToken, refreshToken, errGenTokens := utils.GenerateTokens(user.Id, user.IsSuperAdmin)
	if errGenTokens != nil {
		return nil, errGenTokens
	}
	return &userv1.SignInRes{
		Item: &userv1.User{
			Id:           user.Id.String(),
			Name:         user.Name,
			Email:        user.Email,
			IsSuperAdmin: user.IsSuperAdmin,
			CreatedAt:    user.CreatedAt.Unix(),
			UpdatedAt:    user.UpdatedAt.Unix(),
			Token: &userv1.Token{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			},
		},
	}, nil
}

func (s *learnmarkService) RefreshToken(ctx context.Context, req *userv1.RefreshTokenReq) (*userv1.RefreshTokenRes, error) {
	token, err := utils.ParseToken(req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}
	accessToken, refreshToken, err := utils.GenerateTokens(token["iss"].(uuid.UUID), token["aud"].(bool))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}
	return &userv1.RefreshTokenRes{
		Item: &userv1.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}
