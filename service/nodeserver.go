package service

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/mount-utils"
)

// NodeServer driver
type NodeServer struct {
	Driver  *Driver
	mounter mount.Interface
}

func (s *NodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {

	log.Printf("NodePublishVolume Started - ID: '%s'", req.GetVolumeId())

	if req.GetVolumeId() == "" {
		err := fmt.Errorf("NodePublishVolume error volumeId parameter was empty")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.GetStagingTargetPath() == "" {
		err := fmt.Errorf("NodeUnstageVolume error stagingTargetPath parameter was empty")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.VolumeCapability == nil {
		err := fmt.Errorf("NodeUnstageVolume error volumeCapability parameter was nil")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return nil, nil
}

func (s *NodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {

	slog.Info("NodeUnpublishVolume", "Started - ID", req.GetVolumeId())

	if req.GetTargetPath() == "" {
		err := fmt.Errorf("NodeUnpublishVolume error targetPath parameter was empty")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if req.GetVolumeId() == "" {
		err := fmt.Errorf("NodeUnpublishVolume error volumeId parameter was empty")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	slog.Info("NodeUnpublishVolume", "Finished - ID", req.GetVolumeId())
	return nil, nil
}

func (s *NodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {

	// set as trace because it happens frequently
	slog.Info("NodeGetCapabilities", " Requested - Node", s.Driver.nodeID)

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_UNKNOWN,
					},
				},
			},
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
					},
				},
			},
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
					},
				},
			},
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_VOLUME_MOUNT_GROUP,
					},
				},
			},
		},
	}, nil

}

func (s *NodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {

	slog.Debug("NodeGetInfo", "Requested - Node", s.Driver.nodeID)

	nodeFQDN := getNodeFQDN()
	topo := &csi.Topology{
		Segments: map[string]string{
			"topology.csi.example.com/zone": "true",
		},
	}
	k8sNodeID := nodeFQDN + "$$" + s.Driver.nodeID
	return &csi.NodeGetInfoResponse{
		NodeId:             k8sNodeID,
		AccessibleTopology: topo,
	}, nil
}

func (s NodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	volumeId := req.GetVolumeId()
	slog.Info("NodeStageVolume", "Started - ID", volumeId)

	if volumeId == "" {
		err := fmt.Errorf("NodeStageVolume error volumeId parameter was empty")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if req.VolumeCapability == nil {
		err := fmt.Errorf("NodeStageVolume error volumeCapability parameter was nil")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if req.StagingTargetPath == "" {
		err := fmt.Errorf("NodeStageVolume error stagingTargetPath parameter was empty")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return nil, nil
}

func (s *NodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	volumeId := req.GetVolumeId()

	slog.Info("NodeUnstageVolume", "Started - ID", volumeId)

	if volumeId == "" {
		err := fmt.Errorf("NodeUnstageVolume error volumeId parameter was empty")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if req.StagingTargetPath == "" {
		err := fmt.Errorf("NodeUnstageVolume error stagingTargetPath parameter was empty")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return nil, nil
}

func (s *NodeServer) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, time.Now().String())
}

func (s *NodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	slog.Info("NodeExpandVolume", "Started - ID", req.GetVolumeId())

	if req.GetVolumeId() == "" {
		err := fmt.Errorf("NodeExpandVolume error volumeId parameter was empty")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if req.VolumeCapability == nil {
		err := fmt.Errorf("NodeExpandVolume error volumeCapability parameter was nil")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if req.StagingTargetPath == "" {
		err := fmt.Errorf("NodeExpandVolume error stagingTargetPath parameter was empty")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	slog.Info("NodeExpandVolume", "Finished - ID", req.GetVolumeId())
	response := csi.NodeExpandVolumeResponse{}
	return &response, nil
}

func getNodeFQDN() string {
	cmd := "hostname -f"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		slog.Warn("could not get fqdn with cmd : 'hostname -f', get hostname with 'echo $HOSTNAME'")
		cmd = "echo $HOSTNAME"
		out, err = exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			slog.Error("Failed to execute command", "cmd", cmd)
			return "unknown"
		}
	}
	nodeFQDN := string(out)
	if nodeFQDN == "" {
		slog.Warn("node fqnd not found, setting node name as node fqdn instead")
		nodeFQDN = "unknown"
	}
	nodeFQDN = strings.TrimSuffix(nodeFQDN, "\n")
	return nodeFQDN
}
