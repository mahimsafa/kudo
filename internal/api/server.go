package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/mahimsafa/kudo/internal/api/proto"
	raftlayer "github.com/mahimsafa/kudo/internal/cluster/raft"
	"github.com/mahimsafa/kudo/internal/cluster/state"
)

type Server struct {
	pb.UnimplementedKudoAPIServer
	raft   *raftlayer.RaftNode
	logger *zap.Logger
	grpc   *grpc.Server
}

func NewServer(raft *raftlayer.RaftNode, logger *zap.Logger) *Server {
	return &Server{
		raft:   raft,
		logger: logger,
	}
}

func (s *Server) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	s.grpc = grpc.NewServer()
	pb.RegisterKudoAPIServer(s.grpc, s)

	s.logger.Info("gRPC API server starting", zap.String("addr", addr))
	go s.grpc.Serve(lis)
	return nil
}

func (s *Server) Stop() {
	if s.grpc != nil {
		s.grpc.GracefulStop()
	}
}

func (s *Server) Apply(ctx context.Context, req *pb.ApplyRequest) (*pb.ApplyResponse, error) {
	s.logger.Info("apply request received")

	return &pb.ApplyResponse{
		Success: true,
		Message: "applied successfully",
	}, nil
}

func (s *Server) GetStatus(ctx context.Context, req *pb.StatusRequest) (*pb.StatusResponse, error) {
	fsm := s.raft.FSM()
	app, exists := fsm.GetApplication(req.AppName)
	if !exists {
		return nil, fmt.Errorf("application %q not found", req.AppName)
	}

	instances := fsm.GetInstancesForApp(req.AppName)
	var instanceInfos []*pb.InstanceInfo
	for _, inst := range instances {
		instanceInfos = append(instanceInfos, &pb.InstanceInfo{
			Id:      inst.ID,
			NodeId:  inst.NodeID,
			Status:  inst.Status,
			Address: inst.Address,
		})
	}

	running := 0
	for _, inst := range instances {
		if inst.Status == "running" {
			running++
		}
	}

	return &pb.StatusResponse{
		AppName:         app.Name,
		Adapter:         app.Adapter,
		DesiredReplicas: int32(app.Replicas),
		RunningReplicas: int32(running),
		Instances:       instanceInfos,
	}, nil
}

func (s *Server) ListNodes(ctx context.Context, req *pb.ListNodesRequest) (*pb.ListNodesResponse, error) {
	fsm := s.raft.FSM()
	nodes := fsm.GetAllNodes()

	var nodeInfos []*pb.NodeInfo
	for _, n := range nodes {
		nodeInfos = append(nodeInfos, &pb.NodeInfo{
			Id:      n.ID,
			Name:    n.Name,
			Address: n.Address,
			Status:  n.Status,
		})
	}

	return &pb.ListNodesResponse{Nodes: nodeInfos}, nil
}

func (s *Server) ListApplications(ctx context.Context, req *pb.ListAppsRequest) (*pb.ListAppsResponse, error) {
	fsm := s.raft.FSM()
	apps := fsm.GetAllApplications()

	var appInfos []*pb.AppInfo
	for _, app := range apps {
		appInfos = append(appInfos, &pb.AppInfo{
			Name:     app.Name,
			Adapter:  app.Adapter,
			Replicas: int32(app.Replicas),
		})
	}

	return &pb.ListAppsResponse{Apps: appInfos}, nil
}

func (s *Server) ScaleApplication(ctx context.Context, req *pb.ScaleRequest) (*pb.ScaleResponse, error) {
	fsm := s.raft.FSM()
	app, exists := fsm.GetApplication(req.AppName)
	if !exists {
		return nil, fmt.Errorf("application %q not found", req.AppName)
	}

	app.Replicas = int(req.Replicas)

	cmd := state.Command{
		Op:   state.OpSetApplication,
		Data: mustJSON(app),
	}
	data, _ := json.Marshal(cmd)

	if err := s.raft.Apply(data, 5*time.Second); err != nil {
		return nil, fmt.Errorf("applying scale: %w", err)
	}

	return &pb.ScaleResponse{
		Success: true,
		Message: fmt.Sprintf("scaled %s to %d replicas", req.AppName, req.Replicas),
	}, nil
}

func mustJSON(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
