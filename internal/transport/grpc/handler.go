package grpc

import (
	"context"
	"fmt"

	"github.com/ReilEgor/RepoNotifier/internal/domain/usecase"
	pb "github.com/ReilEgor/RepoNotifier/internal/transport/grpc/proto/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SubscriptionHandler struct {
	pb.UnimplementedSubscriptionServiceServer
	usecase usecase.SubscriptionUseCase
}

func NewSubscriptionHandler(uc usecase.SubscriptionUseCase) *SubscriptionHandler {
	return &SubscriptionHandler{
		usecase: uc,
	}
}

func (h *SubscriptionHandler) Subscribe(ctx context.Context, req *pb.SubscribeRequest) (*pb.SubscribeResponse, error) {
	id, err := h.usecase.Subscribe(ctx, req.GetEmail(), req.GetRepository())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to subscribe: %v", err)
	}

	return &pb.SubscribeResponse{
		Message: fmt.Sprintf("Successfully subscribed %d", id),
		Success: true,
	}, nil
}

func (h *SubscriptionHandler) Unsubscribe(ctx context.Context, req *pb.UnsubscribeRequest) (*pb.UnsubscribeResponse, error) {
	err := h.usecase.Unsubscribe(ctx, req.GetEmail(), req.GetRepository())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unsubscribe: %v", err)
	}

	return &pb.UnsubscribeResponse{
		Message: "Successfully unsubscribed",
		Success: true,
	}, nil
}
