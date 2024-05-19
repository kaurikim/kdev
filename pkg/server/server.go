package server

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"kdev/pkg/service"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

var JwtKey = []byte("my_secret_key")

type ExampleServer struct {
	service.UnimplementedExampleServiceServer
	users  map[string]string
	tokens map[string]string
	mu     sync.Mutex
}

func NewExampleServer() *ExampleServer {
	return &ExampleServer{
		users:  make(map[string]string),
		tokens: make(map[string]string),
	}
}

func (s *ExampleServer) GetExample(ctx context.Context, req *service.ExampleRequest) (*service.ExampleResponse, error) {
	return &service.ExampleResponse{Message: "Hello, " + req.Name}, nil
}

func (s *ExampleServer) Login(ctx context.Context, req *service.LoginRequest) (*service.LoginResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	password, ok := s.users[req.Username]
	if !ok || password != req.Password {
		return nil, errors.New("invalid username or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		return nil, err
	}

	s.tokens[tokenString] = req.Username
	return &service.LoginResponse{Token: tokenString}, nil
}

func (s *ExampleServer) CreateUser(ctx context.Context, req *service.CreateUserRequest) (*service.UserResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[req.Username]; exists {
		return nil, errors.New("user already exists")
	}

	s.users[req.Username] = req.Password
	return &service.UserResponse{Message: "user created successfully"}, nil
}

func (s *ExampleServer) DeleteUser(ctx context.Context, req *service.DeleteUserRequest) (*service.UserResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.users, req.Username)
	return &service.UserResponse{Message: "user deleted successfully"}, nil
}

func (s *ExampleServer) Logout(ctx context.Context, req *service.LogoutRequest) (*service.LogoutResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tokens, req.Token)
	return &service.LogoutResponse{Message: "logout successful"}, nil
}

func UnaryInterceptor(s *ExampleServer) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if info.FullMethod == "/example.ExampleService/Login" || info.FullMethod == "/example.ExampleService/CreateUser" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if tokens, ok := md["authorization"]; ok {
				for _, token := range tokens {
					if strings.HasPrefix(token, "Bearer ") {
						token = strings.TrimPrefix(token, "Bearer ")
						s.mu.Lock()
						_, tokenValid := s.tokens[token]
						s.mu.Unlock()
						if tokenValid {
							return handler(ctx, req)
						}
					}
				}
			}
		}
		return nil, grpc.Errorf(codes.Unauthenticated, "Unauthenticated")
	}
}
