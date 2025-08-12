package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/senyabanana/pvz-service/internal/service"
	pbv1 "github.com/senyabanana/pvz-service/pkg/pb/pvz_v1"
)

type PVZGRPCHandler struct {
	pbv1.UnimplementedPVZServiceServer
	service service.PVZOperations
}

func NewPVZGRPCHandler(service service.PVZOperations) *PVZGRPCHandler {
	return &PVZGRPCHandler{
		service: service,
	}
}

func (h *PVZGRPCHandler) GetPVZList(ctx context.Context, req *pbv1.GetPVZListRequest) (*pbv1.GetPVZListResponse, error) {
	pvzList, err := h.service.GetAllPVZ(ctx)
	if err != nil {
		return nil, err
	}

	var resp pbv1.GetPVZListResponse
	for _, pvz := range pvzList {
		resp.Pvzs = append(resp.Pvzs, &pbv1.PVZ{
			Id:               pvz.ID.String(),
			City:             string(pvz.City),
			RegistrationDate: timestamppb.New(pvz.RegistrationDate),
		})
	}
	return &resp, nil
}
