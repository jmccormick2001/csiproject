package service

import (
	"context"
	"errors"
	"fmt"

	//"infinibox-csi-driver/api"
	//"infinibox-csi-driver/api/clientgo"
	//"infinibox-csi-driver/common"
	//"infinibox-csi-driver/storage"

	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"

	"log/slog"
	//"infinibox-csi-driver/log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ControllerServer controller server setting
type ControllerServer struct {
	Driver *Driver
}

// CreateVolume method create the volume
func (s *ControllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (createVolResp *csi.CreateVolumeResponse, err error) {

	slog.Info("CreateVolume Start", "ID", req.GetName())

	volName := req.GetName()

	reqParameters := req.GetParameters()
	if len(reqParameters) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no Parameters provided to CreateVolume")
	}

	reqCapabilities := req.GetVolumeCapabilities()

	slog.Debug("CreateVolume", "capacity-range", req.GetCapacityRange())
	slog.Debug("CreateVolume", "params", reqParameters)

	// Basic CSI parameter checking across protocols
	if len(volName) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no name provided to CreateVolume")
	}
	if len(reqCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no VolumeCapabilities provided to CreateVolume")
	}
	error := validateCapabilities(reqCapabilities)
	if error != nil {
		return nil, status.Errorf(codes.InvalidArgument, "VolumeCapabilities invalid: %v", error)
	}

	slog.Info("CreateVolume", "Finish - Name", volName, "ID", volName, createVolResp.Volume.VolumeId)
	return
}

// DeleteVolume method delete the volumne
func (s *ControllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (deleteVolResp *csi.DeleteVolumeResponse, err error) {

	volumeId := req.GetVolumeId()

	slog.Info("DeleteVolume", "Start - ID", volumeId)

	if volumeId == "" {
		err := fmt.Errorf("volumeId parameter empty")
		slog.Error("DeleteVolume", "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	slog.Info("DeleteVolume", "Finish - ID", volumeId)
	return
}

// ControllerPublishVolume method
func (s *ControllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (publishVolResp *csi.ControllerPublishVolumeResponse, err error) {

	slog.Info("ControllerPublishVolume", "Start - ID", req.GetVolumeId())

	slog.Debug("ControllerPublishVolume", "ID", req.GetVolumeId(), "nodeID", req.GetNodeId())

	if req.VolumeCapability == nil {
		err = fmt.Errorf("ControllerPublishVolume request VolumeCapability was nil")
		err = status.Errorf(codes.InvalidArgument, err.Error())
		return
	}
	if req.GetVolumeId() == "" {
		err = fmt.Errorf("ControllerPublishVolume request volumeId was empty")
		err = status.Errorf(codes.InvalidArgument, err.Error())
		return
	}

	if req.GetNodeId() == "" {
		err = fmt.Errorf("ControllerPublishVolume request nodeId was empty")
		err = status.Errorf(codes.InvalidArgument, err.Error())
		return
	}

	err = validateNodeID(req.GetNodeId())
	if err != nil {
		return nil, err
	}

	slog.Info("ControllerPublishVolume", "Finish - ID", req.GetVolumeId())

	return
}

// ControllerUnpublishVolume method
func (s *ControllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (unpublishVolResp *csi.ControllerUnpublishVolumeResponse, err error) {
	slog.Info("ControllerUnpublishVolume", "Start - ID", req.GetVolumeId(), "nodeID", req.GetNodeId())

	if req.GetVolumeId() == "" {
		err = fmt.Errorf("ControllerUnpublishVolume request volumeId parameter was empty")
		err = status.Errorf(codes.InvalidArgument, err.Error())
		return
	}

	slog.Info("ControllerUnPublishVolume", "Finish - ID", req.GetVolumeId())

	return
}

func validateCapabilities(capabilities []*csi.VolumeCapability) error {
	isBlock := false
	isFile := false

	if capabilities == nil {
		return errors.New("no volume capabilities specified")
	}

	for _, capability := range capabilities {
		// validate accessMode
		accessMode := capability.GetAccessMode()
		if accessMode == nil {
			return errors.New("no accessmode specified in volume capability")
		}
		mode := accessMode.GetMode()
		// TODO: do something to actually reject invalid access modes, if any
		// there aren't any that we don't support yet, but some combinations are dumb?

		// check block and file behavior
		if block := capability.GetBlock(); block != nil {
			isBlock = true
			if mode == csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER {
				slog.Warn("MULTI_NODE_MULTI_WRITER AccessMode requested for block volume, could be dangerous")
			}
			// TODO: something about SINGLE_NODE_MULTI_WRITER (alpha feature) as well?
		}
		if file := capability.GetMount(); file != nil {
			isFile = true
			// We should validate fs_type and []mount_flags parts of MountVolume message in NFS/TreeQ controllers - CSIC-339
		}
	}

	if isBlock && isFile {
		return errors.New("both file and block volume capabilities specified")
	}

	return nil
}

func (s *ControllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (validateVolCapsResponse *csi.ValidateVolumeCapabilitiesResponse, err error) {
	slog.Info("ValidateVolumeCapabilities Started - ID: %s", req.GetVolumeId())

	if req.GetVolumeId() == "" {
		err := fmt.Errorf("ValidateVolumeCapabilities error volumeId parameter was empty")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if req.VolumeCapabilities == nil {
		err := fmt.Errorf("ValidateVolumeCapabilities error volumeCapabilities parameter was nil")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if len(req.VolumeCapabilities) == 0 {
		err := fmt.Errorf("ValidateVolumeCapabilities error volumeCapabilities parameter was empty")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	slog.Info("ValidateVolumeCapabilities", "Finished - ID", req.GetVolumeId())

	return
}

func (s *ControllerServer) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	slog.Info("ControllerListVolumes Started")

	res := &csi.ListVolumesResponse{
		Entries: make([]*csi.ListVolumesResponse_Entry, 0),
	}

	slog.Info("ControllerListVolumes Finished")

	return res, nil

}

func (s *ControllerServer) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	slog.Info("ControllerListSnapshots Started")

	res := &csi.ListSnapshotsResponse{
		Entries: make([]*csi.ListSnapshotsResponse_Entry, 0),
	}

	slog.Info("ControllerListSnapshots Finished")

	return res, nil
}

func (s *ControllerServer) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *ControllerServer) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: []*csi.ControllerServiceCapability{
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
					},
				},
			},
		},
	}, nil
}

func (s *ControllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (createSnapshotResp *csi.CreateSnapshotResponse, err error) {

	slog.Info("ControllerCreateSnapshot", "Started - ID", req.GetSourceVolumeId())

	slog.Info("ControllerCreateSnapshot", "Finished - ID", req.GetSourceVolumeId())

	return nil, nil
}

func (s *ControllerServer) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (deleteSnapshotResp *csi.DeleteSnapshotResponse, err error) {

	slog.Info("ControllerDeleteSnapshot", "Start - ID", req.GetSnapshotId())

	return nil, nil
}

func (s *ControllerServer) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (expandVolResp *csi.ControllerExpandVolumeResponse, err error) {

	slog.Info("ControllerExpandVolume", "Started - ID", req.GetVolumeId())

	slog.Info("ControllerExpandVolume", "Ended - ID", req.GetVolumeId())

	return
}

func (s *ControllerServer) ControllerModifyVolume(ctx context.Context, req *csi.ControllerModifyVolumeRequest) (resp *csi.ControllerModifyVolumeResponse, err error) {

	slog.Info("ControllerModifyVolume", "Started - ID", req.GetVolumeId())

	slog.Info("ControllerModifyVolume", "Ended - ID", req.GetVolumeId())

	return
}

func (s *ControllerServer) ControllerGetVolume(_ context.Context, _ *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func validateNodeID(nodeID string) error {
	if nodeID == "" {
		return status.Error(codes.InvalidArgument, "node ID empty")
	}
	nodeSplit := strings.Split(nodeID, "$$")
	if len(nodeSplit) != 2 {
		return status.Error(codes.NotFound, "node Id does not follow '<fqdn>$$<id>' pattern")
	}
	return nil
}

// Controller expand volume request validation
func validateExpandVolumeRequest(req *csi.ControllerExpandVolumeRequest) error {
	if req.GetVolumeId() == "" {
		return status.Error(codes.InvalidArgument, "Volume ID cannot be empty")
	}
	capRange := req.GetCapacityRange()
	if capRange == nil {
		return status.Error(codes.InvalidArgument, "CapacityRange cannot be empty")
	}
	return nil
}
