// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        v3.21.12
// source: api/rbac/v1/rbac.proto

package v1

import (
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	v1 "github.com/studio-asd/go-example/proto/types/rbac/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CreateSecurityPermissionRequest struct {
	state           protoimpl.MessageState    `protogen:"open.v1"`
	PermissionName  string                    `protobuf:"bytes,1,opt,name=permission_name,json=permissionName,proto3" json:"permission_name,omitempty"`
	PermissionType  v1.SecurityPermissionType `protobuf:"varint,2,opt,name=permission_type,json=permissionType,proto3,enum=go_example.types.rbac.v1.SecurityPermissionType" json:"permission_type,omitempty"`
	PermissionKey   string                    `protobuf:"bytes,3,opt,name=permission_key,json=permissionKey,proto3" json:"permission_key,omitempty"`
	PermissionValue string                    `protobuf:"bytes,4,opt,name=permission_value,json=permissionValue,proto3" json:"permission_value,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *CreateSecurityPermissionRequest) Reset() {
	*x = CreateSecurityPermissionRequest{}
	mi := &file_api_rbac_v1_rbac_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateSecurityPermissionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateSecurityPermissionRequest) ProtoMessage() {}

func (x *CreateSecurityPermissionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_rbac_v1_rbac_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateSecurityPermissionRequest.ProtoReflect.Descriptor instead.
func (*CreateSecurityPermissionRequest) Descriptor() ([]byte, []int) {
	return file_api_rbac_v1_rbac_proto_rawDescGZIP(), []int{0}
}

func (x *CreateSecurityPermissionRequest) GetPermissionName() string {
	if x != nil {
		return x.PermissionName
	}
	return ""
}

func (x *CreateSecurityPermissionRequest) GetPermissionType() v1.SecurityPermissionType {
	if x != nil {
		return x.PermissionType
	}
	return v1.SecurityPermissionType(0)
}

func (x *CreateSecurityPermissionRequest) GetPermissionKey() string {
	if x != nil {
		return x.PermissionKey
	}
	return ""
}

func (x *CreateSecurityPermissionRequest) GetPermissionValue() string {
	if x != nil {
		return x.PermissionValue
	}
	return ""
}

type CreateSecurityPermissionResponse struct {
	state          protoimpl.MessageState    `protogen:"open.v1"`
	PermissionId   string                    `protobuf:"bytes,1,opt,name=permission_id,json=permissionId,proto3" json:"permission_id,omitempty"`
	PermissionName string                    `protobuf:"bytes,2,opt,name=permission_name,json=permissionName,proto3" json:"permission_name,omitempty"`
	PermissionType v1.SecurityPermissionType `protobuf:"varint,3,opt,name=permission_type,json=permissionType,proto3,enum=go_example.types.rbac.v1.SecurityPermissionType" json:"permission_type,omitempty"`
	CreatedAt      *timestamppb.Timestamp    `protobuf:"bytes,10,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *CreateSecurityPermissionResponse) Reset() {
	*x = CreateSecurityPermissionResponse{}
	mi := &file_api_rbac_v1_rbac_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateSecurityPermissionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateSecurityPermissionResponse) ProtoMessage() {}

func (x *CreateSecurityPermissionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_rbac_v1_rbac_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateSecurityPermissionResponse.ProtoReflect.Descriptor instead.
func (*CreateSecurityPermissionResponse) Descriptor() ([]byte, []int) {
	return file_api_rbac_v1_rbac_proto_rawDescGZIP(), []int{1}
}

func (x *CreateSecurityPermissionResponse) GetPermissionId() string {
	if x != nil {
		return x.PermissionId
	}
	return ""
}

func (x *CreateSecurityPermissionResponse) GetPermissionName() string {
	if x != nil {
		return x.PermissionName
	}
	return ""
}

func (x *CreateSecurityPermissionResponse) GetPermissionType() v1.SecurityPermissionType {
	if x != nil {
		return x.PermissionType
	}
	return v1.SecurityPermissionType(0)
}

func (x *CreateSecurityPermissionResponse) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

type CreateSecurityRoleRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	RoleName      string                 `protobuf:"bytes,1,opt,name=role_name,json=roleName,proto3" json:"role_name,omitempty"`
	PermissionIds []string               `protobuf:"bytes,2,rep,name=permission_ids,json=permissionIds,proto3" json:"permission_ids,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateSecurityRoleRequest) Reset() {
	*x = CreateSecurityRoleRequest{}
	mi := &file_api_rbac_v1_rbac_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateSecurityRoleRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateSecurityRoleRequest) ProtoMessage() {}

func (x *CreateSecurityRoleRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_rbac_v1_rbac_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateSecurityRoleRequest.ProtoReflect.Descriptor instead.
func (*CreateSecurityRoleRequest) Descriptor() ([]byte, []int) {
	return file_api_rbac_v1_rbac_proto_rawDescGZIP(), []int{2}
}

func (x *CreateSecurityRoleRequest) GetRoleName() string {
	if x != nil {
		return x.RoleName
	}
	return ""
}

func (x *CreateSecurityRoleRequest) GetPermissionIds() []string {
	if x != nil {
		return x.PermissionIds
	}
	return nil
}

type CreateSecurityRoleResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	RoleId        string                 `protobuf:"bytes,1,opt,name=role_id,json=roleId,proto3" json:"role_id,omitempty"`
	RoleName      string                 `protobuf:"bytes,2,opt,name=role_name,json=roleName,proto3" json:"role_name,omitempty"`
	CreatedAt     *timestamppb.Timestamp `protobuf:"bytes,10,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateSecurityRoleResponse) Reset() {
	*x = CreateSecurityRoleResponse{}
	mi := &file_api_rbac_v1_rbac_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateSecurityRoleResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateSecurityRoleResponse) ProtoMessage() {}

func (x *CreateSecurityRoleResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_rbac_v1_rbac_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateSecurityRoleResponse.ProtoReflect.Descriptor instead.
func (*CreateSecurityRoleResponse) Descriptor() ([]byte, []int) {
	return file_api_rbac_v1_rbac_proto_rawDescGZIP(), []int{3}
}

func (x *CreateSecurityRoleResponse) GetRoleId() string {
	if x != nil {
		return x.RoleId
	}
	return ""
}

func (x *CreateSecurityRoleResponse) GetRoleName() string {
	if x != nil {
		return x.RoleName
	}
	return ""
}

func (x *CreateSecurityRoleResponse) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

var File_api_rbac_v1_rbac_proto protoreflect.FileDescriptor

var file_api_rbac_v1_rbac_proto_rawDesc = string([]byte{
	0x0a, 0x16, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x62, 0x61, 0x63, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x62,
	0x61, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16, 0x67, 0x6f, 0x5f, 0x65, 0x78, 0x61,
	0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x72, 0x62, 0x61, 0x63, 0x2e, 0x76, 0x31,
	0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x18,
	0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x72, 0x62, 0x61, 0x63, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x62,
	0x61, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8f, 0x02, 0x0a, 0x1f, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x50, 0x65, 0x72, 0x6d, 0x69,
	0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2f, 0x0a, 0x0f,
	0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x06, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x0e, 0x70,
	0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x61, 0x0a,
	0x0f, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x30, 0x2e, 0x67, 0x6f, 0x5f, 0x65, 0x78, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x72, 0x62, 0x61, 0x63, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x50, 0x65, 0x72, 0x6d, 0x69, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x42, 0x06, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01,
	0x52, 0x0e, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x25, 0x0a, 0x0e, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x6b,
	0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x4b, 0x65, 0x79, 0x12, 0x31, 0x0a, 0x10, 0x70, 0x65, 0x72, 0x6d, 0x69,
	0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x06, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x0f, 0x70, 0x65, 0x72, 0x6d, 0x69,
	0x73, 0x73, 0x69, 0x6f, 0x6e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x86, 0x02, 0x0a, 0x20, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x50, 0x65, 0x72,
	0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x23, 0x0a, 0x0d, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x49, 0x64, 0x12, 0x27, 0x0a, 0x0f, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x70,
	0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x59, 0x0a,
	0x0f, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x30, 0x2e, 0x67, 0x6f, 0x5f, 0x65, 0x78, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x72, 0x62, 0x61, 0x63, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x50, 0x65, 0x72, 0x6d, 0x69, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0e, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x12, 0x39, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x64, 0x41, 0x74, 0x22, 0x6f, 0x0a, 0x19, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x63,
	0x75, 0x72, 0x69, 0x74, 0x79, 0x52, 0x6f, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x23, 0x0a, 0x09, 0x72, 0x6f, 0x6c, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x06, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x08, 0x72, 0x6f, 0x6c,
	0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x2d, 0x0a, 0x0e, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73,
	0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x42, 0x06, 0xba,
	0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x0d, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x49, 0x64, 0x73, 0x22, 0x8d, 0x01, 0x0a, 0x1a, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53,
	0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x52, 0x6f, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x72, 0x6f, 0x6c, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x6f, 0x6c, 0x65, 0x49, 0x64, 0x12, 0x1b, 0x0a, 0x09,
	0x72, 0x6f, 0x6c, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x72, 0x6f, 0x6c, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x39, 0x0a, 0x0a, 0x63, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x64, 0x41, 0x74, 0x42, 0x34, 0x5a, 0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x73, 0x74, 0x75, 0x64, 0x69, 0x6f, 0x2d, 0x61, 0x73, 0x64, 0x2f, 0x67, 0x6f,
	0x2d, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x72, 0x62, 0x61, 0x63, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
})

var (
	file_api_rbac_v1_rbac_proto_rawDescOnce sync.Once
	file_api_rbac_v1_rbac_proto_rawDescData []byte
)

func file_api_rbac_v1_rbac_proto_rawDescGZIP() []byte {
	file_api_rbac_v1_rbac_proto_rawDescOnce.Do(func() {
		file_api_rbac_v1_rbac_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_api_rbac_v1_rbac_proto_rawDesc), len(file_api_rbac_v1_rbac_proto_rawDesc)))
	})
	return file_api_rbac_v1_rbac_proto_rawDescData
}

var file_api_rbac_v1_rbac_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_api_rbac_v1_rbac_proto_goTypes = []any{
	(*CreateSecurityPermissionRequest)(nil),  // 0: go_example.api.rbac.v1.CreateSecurityPermissionRequest
	(*CreateSecurityPermissionResponse)(nil), // 1: go_example.api.rbac.v1.CreateSecurityPermissionResponse
	(*CreateSecurityRoleRequest)(nil),        // 2: go_example.api.rbac.v1.CreateSecurityRoleRequest
	(*CreateSecurityRoleResponse)(nil),       // 3: go_example.api.rbac.v1.CreateSecurityRoleResponse
	(v1.SecurityPermissionType)(0),           // 4: go_example.types.rbac.v1.SecurityPermissionType
	(*timestamppb.Timestamp)(nil),            // 5: google.protobuf.Timestamp
}
var file_api_rbac_v1_rbac_proto_depIdxs = []int32{
	4, // 0: go_example.api.rbac.v1.CreateSecurityPermissionRequest.permission_type:type_name -> go_example.types.rbac.v1.SecurityPermissionType
	4, // 1: go_example.api.rbac.v1.CreateSecurityPermissionResponse.permission_type:type_name -> go_example.types.rbac.v1.SecurityPermissionType
	5, // 2: go_example.api.rbac.v1.CreateSecurityPermissionResponse.created_at:type_name -> google.protobuf.Timestamp
	5, // 3: go_example.api.rbac.v1.CreateSecurityRoleResponse.created_at:type_name -> google.protobuf.Timestamp
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_api_rbac_v1_rbac_proto_init() }
func file_api_rbac_v1_rbac_proto_init() {
	if File_api_rbac_v1_rbac_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_api_rbac_v1_rbac_proto_rawDesc), len(file_api_rbac_v1_rbac_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_rbac_v1_rbac_proto_goTypes,
		DependencyIndexes: file_api_rbac_v1_rbac_proto_depIdxs,
		MessageInfos:      file_api_rbac_v1_rbac_proto_msgTypes,
	}.Build()
	File_api_rbac_v1_rbac_proto = out.File
	file_api_rbac_v1_rbac_proto_goTypes = nil
	file_api_rbac_v1_rbac_proto_depIdxs = nil
}
