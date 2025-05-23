// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        v3.21.12
// source: api/ledger/v1/currency.proto

package v1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
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

type CurrencyListResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Currencies    []*Currency            `protobuf:"bytes,1,rep,name=currencies,proto3" json:"currencies,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CurrencyListResponse) Reset() {
	*x = CurrencyListResponse{}
	mi := &file_api_ledger_v1_currency_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CurrencyListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CurrencyListResponse) ProtoMessage() {}

func (x *CurrencyListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_ledger_v1_currency_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CurrencyListResponse.ProtoReflect.Descriptor instead.
func (*CurrencyListResponse) Descriptor() ([]byte, []int) {
	return file_api_ledger_v1_currency_proto_rawDescGZIP(), []int{0}
}

func (x *CurrencyListResponse) GetCurrencies() []*Currency {
	if x != nil {
		return x.Currencies
	}
	return nil
}

type Currency struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int32                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name          string                 `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Exponent      int32                  `protobuf:"varint,3,opt,name=exponent,proto3" json:"exponent,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Currency) Reset() {
	*x = Currency{}
	mi := &file_api_ledger_v1_currency_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Currency) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Currency) ProtoMessage() {}

func (x *Currency) ProtoReflect() protoreflect.Message {
	mi := &file_api_ledger_v1_currency_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Currency.ProtoReflect.Descriptor instead.
func (*Currency) Descriptor() ([]byte, []int) {
	return file_api_ledger_v1_currency_proto_rawDescGZIP(), []int{1}
}

func (x *Currency) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Currency) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Currency) GetExponent() int32 {
	if x != nil {
		return x.Exponent
	}
	return 0
}

var File_api_ledger_v1_currency_proto protoreflect.FileDescriptor

var file_api_ledger_v1_currency_proto_rawDesc = string([]byte{
	0x0a, 0x1c, 0x61, 0x70, 0x69, 0x2f, 0x6c, 0x65, 0x64, 0x67, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f,
	0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x18,
	0x67, 0x6f, 0x5f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x6c,
	0x65, 0x64, 0x67, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x22, 0x5a, 0x0a, 0x14, 0x43, 0x75, 0x72, 0x72,
	0x65, 0x6e, 0x63, 0x79, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x42, 0x0a, 0x0a, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x67, 0x6f, 0x5f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c,
	0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x6c, 0x65, 0x64, 0x67, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e,
	0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79, 0x52, 0x0a, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e,
	0x63, 0x69, 0x65, 0x73, 0x22, 0x4a, 0x0a, 0x08, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x65, 0x78, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x65, 0x78, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74,
	0x42, 0x36, 0x5a, 0x34, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73,
	0x74, 0x75, 0x64, 0x69, 0x6f, 0x2d, 0x61, 0x73, 0x64, 0x2f, 0x67, 0x6f, 0x2d, 0x65, 0x78, 0x61,
	0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x6c,
	0x65, 0x64, 0x67, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_api_ledger_v1_currency_proto_rawDescOnce sync.Once
	file_api_ledger_v1_currency_proto_rawDescData []byte
)

func file_api_ledger_v1_currency_proto_rawDescGZIP() []byte {
	file_api_ledger_v1_currency_proto_rawDescOnce.Do(func() {
		file_api_ledger_v1_currency_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_api_ledger_v1_currency_proto_rawDesc), len(file_api_ledger_v1_currency_proto_rawDesc)))
	})
	return file_api_ledger_v1_currency_proto_rawDescData
}

var file_api_ledger_v1_currency_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_ledger_v1_currency_proto_goTypes = []any{
	(*CurrencyListResponse)(nil), // 0: go_example.api.ledger.v1.CurrencyListResponse
	(*Currency)(nil),             // 1: go_example.api.ledger.v1.Currency
}
var file_api_ledger_v1_currency_proto_depIdxs = []int32{
	1, // 0: go_example.api.ledger.v1.CurrencyListResponse.currencies:type_name -> go_example.api.ledger.v1.Currency
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_ledger_v1_currency_proto_init() }
func file_api_ledger_v1_currency_proto_init() {
	if File_api_ledger_v1_currency_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_api_ledger_v1_currency_proto_rawDesc), len(file_api_ledger_v1_currency_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_ledger_v1_currency_proto_goTypes,
		DependencyIndexes: file_api_ledger_v1_currency_proto_depIdxs,
		MessageInfos:      file_api_ledger_v1_currency_proto_msgTypes,
	}.Build()
	File_api_ledger_v1_currency_proto = out.File
	file_api_ledger_v1_currency_proto_goTypes = nil
	file_api_ledger_v1_currency_proto_depIdxs = nil
}
