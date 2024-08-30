// Code generated by protoc-gen-go. DO NOT EDIT.
// source: internal/scheduling/protobuf/solver.proto

// package name for the buffer will be used later

package serverledge

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// argument
type Request struct {
	Policy                  *string     `protobuf:"bytes,1,req,name=policy" json:"policy,omitempty"`
	OffloadLatencyCloud     *float32    `protobuf:"fixed32,2,req,name=offload_latency_cloud,json=offloadLatencyCloud" json:"offload_latency_cloud,omitempty"`
	OffloadLatencyEdge      *float32    `protobuf:"fixed32,3,req,name=offload_latency_edge,json=offloadLatencyEdge" json:"offload_latency_edge,omitempty"`
	Functions               []*Function `protobuf:"bytes,4,rep,name=functions" json:"functions,omitempty"`
	Classes                 []*QosClass `protobuf:"bytes,5,rep,name=classes" json:"classes,omitempty"`
	CostCloud               *float32    `protobuf:"fixed32,6,req,name=cost_cloud,json=costCloud" json:"cost_cloud,omitempty"`
	LocalBudget             *float32    `protobuf:"fixed32,7,req,name=local_budget,json=localBudget" json:"local_budget,omitempty"`
	MemoryLocal             *float32    `protobuf:"fixed32,8,req,name=memory_local,json=memoryLocal" json:"memory_local,omitempty"`
	CpuLocal                *float32    `protobuf:"fixed32,9,req,name=cpu_local,json=cpuLocal" json:"cpu_local,omitempty"`
	MemoryAggregate         *float32    `protobuf:"fixed32,10,req,name=memory_aggregate,json=memoryAggregate" json:"memory_aggregate,omitempty"`
	UsableMemoryCoefficient *float32    `protobuf:"fixed32,11,req,name=usable_memory_coefficient,json=usableMemoryCoefficient" json:"usable_memory_coefficient,omitempty"`
	XXX_NoUnkeyedLiteral    struct{}    `json:"-"`
	XXX_unrecognized        []byte      `json:"-"`
	XXX_sizecache           int32       `json:"-"`
}

func (m *Request) Reset()         { *m = Request{} }
func (m *Request) String() string { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()    {}
func (*Request) Descriptor() ([]byte, []int) {
	return fileDescriptor_6199495c0a4eff57, []int{0}
}

func (m *Request) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Request.Unmarshal(m, b)
}
func (m *Request) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Request.Marshal(b, m, deterministic)
}
func (m *Request) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Request.Merge(m, src)
}
func (m *Request) XXX_Size() int {
	return xxx_messageInfo_Request.Size(m)
}
func (m *Request) XXX_DiscardUnknown() {
	xxx_messageInfo_Request.DiscardUnknown(m)
}

var xxx_messageInfo_Request proto.InternalMessageInfo

func (m *Request) GetPolicy() string {
	if m != nil && m.Policy != nil {
		return *m.Policy
	}
	return ""
}

func (m *Request) GetOffloadLatencyCloud() float32 {
	if m != nil && m.OffloadLatencyCloud != nil {
		return *m.OffloadLatencyCloud
	}
	return 0
}

func (m *Request) GetOffloadLatencyEdge() float32 {
	if m != nil && m.OffloadLatencyEdge != nil {
		return *m.OffloadLatencyEdge
	}
	return 0
}

func (m *Request) GetFunctions() []*Function {
	if m != nil {
		return m.Functions
	}
	return nil
}

func (m *Request) GetClasses() []*QosClass {
	if m != nil {
		return m.Classes
	}
	return nil
}

func (m *Request) GetCostCloud() float32 {
	if m != nil && m.CostCloud != nil {
		return *m.CostCloud
	}
	return 0
}

func (m *Request) GetLocalBudget() float32 {
	if m != nil && m.LocalBudget != nil {
		return *m.LocalBudget
	}
	return 0
}

func (m *Request) GetMemoryLocal() float32 {
	if m != nil && m.MemoryLocal != nil {
		return *m.MemoryLocal
	}
	return 0
}

func (m *Request) GetCpuLocal() float32 {
	if m != nil && m.CpuLocal != nil {
		return *m.CpuLocal
	}
	return 0
}

func (m *Request) GetMemoryAggregate() float32 {
	if m != nil && m.MemoryAggregate != nil {
		return *m.MemoryAggregate
	}
	return 0
}

func (m *Request) GetUsableMemoryCoefficient() float32 {
	if m != nil && m.UsableMemoryCoefficient != nil {
		return *m.UsableMemoryCoefficient
	}
	return 0
}

type FunctionInvocation struct {
	QosClass             *string  `protobuf:"bytes,1,req,name=qos_class,json=qosClass" json:"qos_class,omitempty"`
	Arrivals             *float32 `protobuf:"fixed32,2,req,name=arrivals" json:"arrivals,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FunctionInvocation) Reset()         { *m = FunctionInvocation{} }
func (m *FunctionInvocation) String() string { return proto.CompactTextString(m) }
func (*FunctionInvocation) ProtoMessage()    {}
func (*FunctionInvocation) Descriptor() ([]byte, []int) {
	return fileDescriptor_6199495c0a4eff57, []int{1}
}

func (m *FunctionInvocation) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FunctionInvocation.Unmarshal(m, b)
}
func (m *FunctionInvocation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FunctionInvocation.Marshal(b, m, deterministic)
}
func (m *FunctionInvocation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FunctionInvocation.Merge(m, src)
}
func (m *FunctionInvocation) XXX_Size() int {
	return xxx_messageInfo_FunctionInvocation.Size(m)
}
func (m *FunctionInvocation) XXX_DiscardUnknown() {
	xxx_messageInfo_FunctionInvocation.DiscardUnknown(m)
}

var xxx_messageInfo_FunctionInvocation proto.InternalMessageInfo

func (m *FunctionInvocation) GetQosClass() string {
	if m != nil && m.QosClass != nil {
		return *m.QosClass
	}
	return ""
}

func (m *FunctionInvocation) GetArrivals() float32 {
	if m != nil && m.Arrivals != nil {
		return *m.Arrivals
	}
	return 0
}

type Function struct {
	Name                   *string               `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Memory                 *int32                `protobuf:"varint,2,req,name=memory" json:"memory,omitempty"`
	Cpu                    *float32              `protobuf:"fixed32,3,req,name=cpu" json:"cpu,omitempty"`
	Invocations            []*FunctionInvocation `protobuf:"bytes,4,rep,name=invocations" json:"invocations,omitempty"`
	Duration               *float32              `protobuf:"fixed32,5,req,name=duration" json:"duration,omitempty"`
	DurationOffloadedCloud *float32              `protobuf:"fixed32,6,req,name=duration_offloaded_cloud,json=durationOffloadedCloud" json:"duration_offloaded_cloud,omitempty"`
	DurationOffloadedEdge  *float32              `protobuf:"fixed32,7,req,name=duration_offloaded_edge,json=durationOffloadedEdge" json:"duration_offloaded_edge,omitempty"`
	InitTime               *float32              `protobuf:"fixed32,8,req,name=init_time,json=initTime" json:"init_time,omitempty"`
	InitTimeOffloadedCloud *float32              `protobuf:"fixed32,9,req,name=init_time_offloaded_cloud,json=initTimeOffloadedCloud" json:"init_time_offloaded_cloud,omitempty"`
	InitTimeOffloadedEdge  *float32              `protobuf:"fixed32,10,req,name=init_time_offloaded_edge,json=initTimeOffloadedEdge" json:"init_time_offloaded_edge,omitempty"`
	Pcold                  *float32              `protobuf:"fixed32,11,req,name=pcold" json:"pcold,omitempty"`
	PcoldOffloadedCloud    *float32              `protobuf:"fixed32,12,req,name=pcold_offloaded_cloud,json=pcoldOffloadedCloud" json:"pcold_offloaded_cloud,omitempty"`
	PcoldOffloadedEdge     *float32              `protobuf:"fixed32,13,req,name=pcold_offloaded_edge,json=pcoldOffloadedEdge" json:"pcold_offloaded_edge,omitempty"`
	BandwidthCloud         *float32              `protobuf:"fixed32,14,req,name=bandwidth_cloud,json=bandwidthCloud" json:"bandwidth_cloud,omitempty"`
	BandwidthEdge          *float32              `protobuf:"fixed32,15,req,name=bandwidth_edge,json=bandwidthEdge" json:"bandwidth_edge,omitempty"`
	InputSize              *float32              `protobuf:"fixed32,16,req,name=input_size,json=inputSize" json:"input_size,omitempty"`
	XXX_NoUnkeyedLiteral   struct{}              `json:"-"`
	XXX_unrecognized       []byte                `json:"-"`
	XXX_sizecache          int32                 `json:"-"`
}

func (m *Function) Reset()         { *m = Function{} }
func (m *Function) String() string { return proto.CompactTextString(m) }
func (*Function) ProtoMessage()    {}
func (*Function) Descriptor() ([]byte, []int) {
	return fileDescriptor_6199495c0a4eff57, []int{2}
}

func (m *Function) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Function.Unmarshal(m, b)
}
func (m *Function) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Function.Marshal(b, m, deterministic)
}
func (m *Function) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Function.Merge(m, src)
}
func (m *Function) XXX_Size() int {
	return xxx_messageInfo_Function.Size(m)
}
func (m *Function) XXX_DiscardUnknown() {
	xxx_messageInfo_Function.DiscardUnknown(m)
}

var xxx_messageInfo_Function proto.InternalMessageInfo

func (m *Function) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *Function) GetMemory() int32 {
	if m != nil && m.Memory != nil {
		return *m.Memory
	}
	return 0
}

func (m *Function) GetCpu() float32 {
	if m != nil && m.Cpu != nil {
		return *m.Cpu
	}
	return 0
}

func (m *Function) GetInvocations() []*FunctionInvocation {
	if m != nil {
		return m.Invocations
	}
	return nil
}

func (m *Function) GetDuration() float32 {
	if m != nil && m.Duration != nil {
		return *m.Duration
	}
	return 0
}

func (m *Function) GetDurationOffloadedCloud() float32 {
	if m != nil && m.DurationOffloadedCloud != nil {
		return *m.DurationOffloadedCloud
	}
	return 0
}

func (m *Function) GetDurationOffloadedEdge() float32 {
	if m != nil && m.DurationOffloadedEdge != nil {
		return *m.DurationOffloadedEdge
	}
	return 0
}

func (m *Function) GetInitTime() float32 {
	if m != nil && m.InitTime != nil {
		return *m.InitTime
	}
	return 0
}

func (m *Function) GetInitTimeOffloadedCloud() float32 {
	if m != nil && m.InitTimeOffloadedCloud != nil {
		return *m.InitTimeOffloadedCloud
	}
	return 0
}

func (m *Function) GetInitTimeOffloadedEdge() float32 {
	if m != nil && m.InitTimeOffloadedEdge != nil {
		return *m.InitTimeOffloadedEdge
	}
	return 0
}

func (m *Function) GetPcold() float32 {
	if m != nil && m.Pcold != nil {
		return *m.Pcold
	}
	return 0
}

func (m *Function) GetPcoldOffloadedCloud() float32 {
	if m != nil && m.PcoldOffloadedCloud != nil {
		return *m.PcoldOffloadedCloud
	}
	return 0
}

func (m *Function) GetPcoldOffloadedEdge() float32 {
	if m != nil && m.PcoldOffloadedEdge != nil {
		return *m.PcoldOffloadedEdge
	}
	return 0
}

func (m *Function) GetBandwidthCloud() float32 {
	if m != nil && m.BandwidthCloud != nil {
		return *m.BandwidthCloud
	}
	return 0
}

func (m *Function) GetBandwidthEdge() float32 {
	if m != nil && m.BandwidthEdge != nil {
		return *m.BandwidthEdge
	}
	return 0
}

func (m *Function) GetInputSize() float32 {
	if m != nil && m.InputSize != nil {
		return *m.InputSize
	}
	return 0
}

type QosClass struct {
	Name                 *string  `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Utility              *float32 `protobuf:"fixed32,2,req,name=utility" json:"utility,omitempty"`
	MaxResponseTime      *float32 `protobuf:"fixed32,3,req,name=max_response_time,json=maxResponseTime" json:"max_response_time,omitempty"`
	CompletedPercentage  *float32 `protobuf:"fixed32,4,req,name=completed_percentage,json=completedPercentage" json:"completed_percentage,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *QosClass) Reset()         { *m = QosClass{} }
func (m *QosClass) String() string { return proto.CompactTextString(m) }
func (*QosClass) ProtoMessage()    {}
func (*QosClass) Descriptor() ([]byte, []int) {
	return fileDescriptor_6199495c0a4eff57, []int{3}
}

func (m *QosClass) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QosClass.Unmarshal(m, b)
}
func (m *QosClass) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QosClass.Marshal(b, m, deterministic)
}
func (m *QosClass) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QosClass.Merge(m, src)
}
func (m *QosClass) XXX_Size() int {
	return xxx_messageInfo_QosClass.Size(m)
}
func (m *QosClass) XXX_DiscardUnknown() {
	xxx_messageInfo_QosClass.DiscardUnknown(m)
}

var xxx_messageInfo_QosClass proto.InternalMessageInfo

func (m *QosClass) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *QosClass) GetUtility() float32 {
	if m != nil && m.Utility != nil {
		return *m.Utility
	}
	return 0
}

func (m *QosClass) GetMaxResponseTime() float32 {
	if m != nil && m.MaxResponseTime != nil {
		return *m.MaxResponseTime
	}
	return 0
}

func (m *QosClass) GetCompletedPercentage() float32 {
	if m != nil && m.CompletedPercentage != nil {
		return *m.CompletedPercentage
	}
	return 0
}

type ClassResponse struct {
	Name                 *string  `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	PL                   *float32 `protobuf:"fixed32,2,req,name=pL" json:"pL,omitempty"`
	PC                   *float32 `protobuf:"fixed32,3,req,name=pC" json:"pC,omitempty"`
	PE                   *float32 `protobuf:"fixed32,4,req,name=pE" json:"pE,omitempty"`
	PD                   *float32 `protobuf:"fixed32,5,req,name=pD" json:"pD,omitempty"`
	Share                *float32 `protobuf:"fixed32,6,req,name=share" json:"share,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ClassResponse) Reset()         { *m = ClassResponse{} }
func (m *ClassResponse) String() string { return proto.CompactTextString(m) }
func (*ClassResponse) ProtoMessage()    {}
func (*ClassResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_6199495c0a4eff57, []int{4}
}

func (m *ClassResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ClassResponse.Unmarshal(m, b)
}
func (m *ClassResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ClassResponse.Marshal(b, m, deterministic)
}
func (m *ClassResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ClassResponse.Merge(m, src)
}
func (m *ClassResponse) XXX_Size() int {
	return xxx_messageInfo_ClassResponse.Size(m)
}
func (m *ClassResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ClassResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ClassResponse proto.InternalMessageInfo

func (m *ClassResponse) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *ClassResponse) GetPL() float32 {
	if m != nil && m.PL != nil {
		return *m.PL
	}
	return 0
}

func (m *ClassResponse) GetPC() float32 {
	if m != nil && m.PC != nil {
		return *m.PC
	}
	return 0
}

func (m *ClassResponse) GetPE() float32 {
	if m != nil && m.PE != nil {
		return *m.PE
	}
	return 0
}

func (m *ClassResponse) GetPD() float32 {
	if m != nil && m.PD != nil {
		return *m.PD
	}
	return 0
}

func (m *ClassResponse) GetShare() float32 {
	if m != nil && m.Share != nil {
		return *m.Share
	}
	return 0
}

type FunctionResponse struct {
	Name                 *string          `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	ClassResponses       []*ClassResponse `protobuf:"bytes,2,rep,name=class_responses,json=classResponses" json:"class_responses,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *FunctionResponse) Reset()         { *m = FunctionResponse{} }
func (m *FunctionResponse) String() string { return proto.CompactTextString(m) }
func (*FunctionResponse) ProtoMessage()    {}
func (*FunctionResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_6199495c0a4eff57, []int{5}
}

func (m *FunctionResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FunctionResponse.Unmarshal(m, b)
}
func (m *FunctionResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FunctionResponse.Marshal(b, m, deterministic)
}
func (m *FunctionResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FunctionResponse.Merge(m, src)
}
func (m *FunctionResponse) XXX_Size() int {
	return xxx_messageInfo_FunctionResponse.Size(m)
}
func (m *FunctionResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_FunctionResponse.DiscardUnknown(m)
}

var xxx_messageInfo_FunctionResponse proto.InternalMessageInfo

func (m *FunctionResponse) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *FunctionResponse) GetClassResponses() []*ClassResponse {
	if m != nil {
		return m.ClassResponses
	}
	return nil
}

type Response struct {
	FResponse            []*FunctionResponse `protobuf:"bytes,1,rep,name=f_response,json=fResponse" json:"f_response,omitempty"`
	TimeTaken            *float32            `protobuf:"fixed32,2,req,name=time_taken,json=timeTaken" json:"time_taken,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *Response) Reset()         { *m = Response{} }
func (m *Response) String() string { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()    {}
func (*Response) Descriptor() ([]byte, []int) {
	return fileDescriptor_6199495c0a4eff57, []int{6}
}

func (m *Response) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Response.Unmarshal(m, b)
}
func (m *Response) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Response.Marshal(b, m, deterministic)
}
func (m *Response) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Response.Merge(m, src)
}
func (m *Response) XXX_Size() int {
	return xxx_messageInfo_Response.Size(m)
}
func (m *Response) XXX_DiscardUnknown() {
	xxx_messageInfo_Response.DiscardUnknown(m)
}

var xxx_messageInfo_Response proto.InternalMessageInfo

func (m *Response) GetFResponse() []*FunctionResponse {
	if m != nil {
		return m.FResponse
	}
	return nil
}

func (m *Response) GetTimeTaken() float32 {
	if m != nil && m.TimeTaken != nil {
		return *m.TimeTaken
	}
	return 0
}

func init() {
	proto.RegisterType((*Request)(nil), "solver.Request")
	proto.RegisterType((*FunctionInvocation)(nil), "solver.FunctionInvocation")
	proto.RegisterType((*Function)(nil), "solver.Function")
	proto.RegisterType((*QosClass)(nil), "solver.QosClass")
	proto.RegisterType((*ClassResponse)(nil), "solver.ClassResponse")
	proto.RegisterType((*FunctionResponse)(nil), "solver.FunctionResponse")
	proto.RegisterType((*Response)(nil), "solver.Response")
}

func init() {
	proto.RegisterFile("internal/scheduling/protobuf/solver.proto", fileDescriptor_6199495c0a4eff57)
}

var fileDescriptor_6199495c0a4eff57 = []byte{
	// 833 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x55, 0x51, 0x6f, 0x1b, 0x45,
	0x10, 0x56, 0xec, 0x38, 0xb1, 0x27, 0x4d, 0x6c, 0x96, 0x24, 0xdd, 0xa6, 0x8a, 0x14, 0x2c, 0x10,
	0x69, 0x1e, 0xe2, 0x12, 0x21, 0x0a, 0x08, 0x21, 0x51, 0x37, 0x48, 0x48, 0xa9, 0x80, 0x6b, 0x9f,
	0x78, 0x39, 0xad, 0xf7, 0xe6, 0x2e, 0x2b, 0xce, 0xb7, 0x97, 0xdb, 0xbd, 0x50, 0xf7, 0x57, 0xf0,
	0xc6, 0x9f, 0xe2, 0x47, 0xa1, 0x9d, 0xdd, 0x3b, 0xc7, 0x76, 0xc4, 0x4b, 0xb4, 0xf3, 0x7d, 0x33,
	0x73, 0x5f, 0x76, 0xbe, 0x59, 0xc3, 0x0b, 0x55, 0x58, 0xac, 0x0a, 0x91, 0x4f, 0x8c, 0xbc, 0xc5,
	0xa4, 0xce, 0x55, 0x91, 0x4d, 0xca, 0x4a, 0x5b, 0x3d, 0xab, 0xd3, 0x89, 0xd1, 0xf9, 0x3d, 0x56,
	0x97, 0x14, 0xb3, 0x1d, 0x1f, 0x8d, 0xff, 0xed, 0xc2, 0x6e, 0x84, 0x77, 0x35, 0x1a, 0xcb, 0x8e,
	0x61, 0xa7, 0xd4, 0xb9, 0x92, 0x0b, 0xbe, 0x75, 0xd6, 0x39, 0x1f, 0x44, 0x21, 0x62, 0x57, 0x70,
	0xa4, 0xd3, 0x34, 0xd7, 0x22, 0x89, 0x73, 0x61, 0xb1, 0x90, 0x8b, 0x58, 0xe6, 0xba, 0x4e, 0x78,
	0xe7, 0xac, 0x73, 0xde, 0x89, 0x3e, 0x0d, 0xe4, 0x8d, 0xe7, 0xa6, 0x8e, 0x62, 0x2f, 0xe1, 0x70,
	0xbd, 0x06, 0x93, 0x0c, 0x79, 0x97, 0x4a, 0xd8, 0x6a, 0xc9, 0x75, 0x92, 0x21, 0xbb, 0x84, 0x41,
	0x5a, 0x17, 0xd2, 0x2a, 0x5d, 0x18, 0xbe, 0x7d, 0xd6, 0x3d, 0xdf, 0xbb, 0x1a, 0x5d, 0x06, 0xcd,
	0x3f, 0x07, 0x22, 0x5a, 0xa6, 0xb0, 0x0b, 0xd8, 0x95, 0xb9, 0x30, 0x06, 0x0d, 0xef, 0xad, 0x66,
	0xff, 0xae, 0xcd, 0xd4, 0x31, 0x51, 0x93, 0xc0, 0x4e, 0x01, 0xa4, 0x36, 0x36, 0xc8, 0xde, 0x21,
	0x0d, 0x03, 0x87, 0x78, 0xb1, 0x9f, 0xc1, 0x93, 0x5c, 0x4b, 0x91, 0xc7, 0xb3, 0x3a, 0xc9, 0xd0,
	0xf2, 0x5d, 0x4a, 0xd8, 0x23, 0xec, 0x35, 0x41, 0x2e, 0x65, 0x8e, 0x73, 0x5d, 0x2d, 0x62, 0x42,
	0x79, 0xdf, 0xa7, 0x78, 0xec, 0xc6, 0x41, 0xec, 0x39, 0x0c, 0x64, 0x59, 0x07, 0x7e, 0x40, 0x7c,
	0x5f, 0x96, 0xb5, 0x27, 0x5f, 0xc0, 0x28, 0xd4, 0x8b, 0x2c, 0xab, 0x30, 0x13, 0x16, 0x39, 0x50,
	0xce, 0xd0, 0xe3, 0x3f, 0x35, 0x30, 0xfb, 0x1e, 0x9e, 0xd5, 0x46, 0xcc, 0x72, 0x8c, 0x43, 0x85,
	0xd4, 0x98, 0xa6, 0x4a, 0x2a, 0x2c, 0x2c, 0xdf, 0xa3, 0x9a, 0xa7, 0x3e, 0xe1, 0x2d, 0xf1, 0xd3,
	0x25, 0x3d, 0x7e, 0x0b, 0xac, 0xb9, 0xab, 0x5f, 0x8a, 0x7b, 0x2d, 0x85, 0x3b, 0x39, 0x65, 0x77,
	0xda, 0xc4, 0x74, 0x1b, 0x61, 0xb6, 0xfd, 0xbb, 0x70, 0x49, 0xec, 0x04, 0xfa, 0xa2, 0xaa, 0xd4,
	0xbd, 0xc8, 0x4d, 0x18, 0x68, 0x1b, 0x8f, 0xff, 0xee, 0x41, 0xbf, 0xe9, 0xc7, 0x18, 0x6c, 0x17,
	0x62, 0x8e, 0xa1, 0x01, 0x9d, 0x9d, 0x65, 0xbc, 0x48, 0x2a, 0xed, 0x45, 0x21, 0x62, 0x23, 0xe8,
	0xca, 0xb2, 0x0e, 0xd3, 0x76, 0x47, 0xf6, 0x03, 0xec, 0xa9, 0x56, 0x51, 0x33, 0xe0, 0x93, 0xf5,
	0x01, 0x2f, 0x45, 0x47, 0x0f, 0xd3, 0x9d, 0xc8, 0xa4, 0xae, 0x28, 0xe0, 0x3d, 0x2f, 0xb2, 0x89,
	0xd9, 0xb7, 0xc0, 0x9b, 0x73, 0x1c, 0x7c, 0x85, 0xc9, 0xca, 0xa8, 0x8f, 0x1b, 0xfe, 0xd7, 0x86,
	0xf6, 0x73, 0xff, 0x06, 0x9e, 0x3e, 0x52, 0x49, 0x3e, 0xf5, 0x16, 0x38, 0xda, 0x28, 0x24, 0xab,
	0x3e, 0x87, 0x81, 0x2a, 0x94, 0x8d, 0xad, 0x9a, 0x63, 0x70, 0x42, 0xdf, 0x01, 0xef, 0xd5, 0x1c,
	0xd9, 0x77, 0xf0, 0xac, 0x25, 0x37, 0xf4, 0x78, 0x5b, 0x1c, 0x37, 0xc9, 0x6b, 0x7a, 0x5e, 0x01,
	0x7f, 0xac, 0x94, 0x04, 0x79, 0xb3, 0x1c, 0x6d, 0x54, 0x92, 0xa0, 0x43, 0xe8, 0x95, 0x52, 0xe7,
	0x49, 0xb0, 0x87, 0x0f, 0xdc, 0xde, 0xd2, 0x61, 0x43, 0xc5, 0x13, 0xbf, 0xb7, 0x44, 0xae, 0x49,
	0x78, 0x09, 0x87, 0xeb, 0x35, 0xf4, 0xf9, 0x7d, 0xbf, 0xb7, 0xab, 0x25, 0xf4, 0xed, 0x2f, 0x61,
	0x38, 0x13, 0x45, 0xf2, 0x97, 0x4a, 0xec, 0x6d, 0xe8, 0x7f, 0x40, 0xc9, 0x07, 0x2d, 0xec, 0x5b,
	0x7f, 0x01, 0x4b, 0xc4, 0x37, 0x1d, 0x52, 0xde, 0x7e, 0x8b, 0x52, 0xbf, 0x53, 0x00, 0x55, 0x94,
	0xb5, 0x8d, 0x8d, 0xfa, 0x88, 0x7c, 0xe4, 0x77, 0x95, 0x90, 0x77, 0xea, 0x23, 0x8e, 0xff, 0xd9,
	0x82, 0x7e, 0xb3, 0xe0, 0x8f, 0x5a, 0x92, 0xc3, 0x6e, 0x6d, 0x55, 0xae, 0xec, 0x22, 0xd8, 0xb9,
	0x09, 0xd9, 0x05, 0x7c, 0x32, 0x17, 0x1f, 0xe2, 0x0a, 0x4d, 0xa9, 0x0b, 0x83, 0x7e, 0x7c, 0xdd,
	0xb0, 0x84, 0xe2, 0x43, 0x14, 0x70, 0x9a, 0xe2, 0x57, 0x70, 0x28, 0xf5, 0xbc, 0xcc, 0xd1, 0x62,
	0x12, 0x97, 0x58, 0x49, 0x2c, 0xac, 0xc8, 0x90, 0x6f, 0xfb, 0xab, 0x6b, 0xb9, 0xdf, 0x5a, 0x6a,
	0xbc, 0x80, 0x7d, 0xff, 0xec, 0x84, 0x3e, 0x8f, 0xaa, 0x3b, 0x80, 0x4e, 0x79, 0x13, 0x84, 0x75,
	0xca, 0x1b, 0x8a, 0xa7, 0x41, 0x44, 0xa7, 0x9c, 0x52, 0x7c, 0x1d, 0xbe, 0xd2, 0x29, 0xaf, 0x29,
	0x7e, 0x13, 0x2c, 0xdf, 0x29, 0xdf, 0xb8, 0x49, 0x9b, 0x5b, 0x51, 0x61, 0x70, 0xb6, 0x0f, 0xc6,
	0x29, 0x8c, 0xda, 0x27, 0xf2, 0xff, 0xbe, 0xfe, 0x23, 0x0c, 0xe9, 0x11, 0x68, 0xef, 0xc0, 0xad,
	0xbc, 0x5b, 0xc4, 0xa3, 0x66, 0x11, 0x57, 0xfe, 0x83, 0xe8, 0x40, 0x3e, 0x0c, 0xcd, 0x78, 0x06,
	0xfd, 0xb6, 0xff, 0x2b, 0x80, 0xb4, 0xed, 0xc3, 0xb7, 0xa8, 0x0d, 0xdf, 0x78, 0xb0, 0x9b, 0x4e,
	0x83, 0xb4, 0x2d, 0x3c, 0x05, 0x20, 0x83, 0x5b, 0xf1, 0x27, 0x16, 0xe1, 0x2a, 0x06, 0x0e, 0x79,
	0xef, 0x80, 0xab, 0xaf, 0x61, 0xe7, 0x1d, 0x35, 0x61, 0x17, 0xd0, 0xa3, 0x13, 0x1b, 0x36, 0x6d,
	0xc3, 0x2f, 0xd5, 0xc9, 0x68, 0x09, 0xf8, 0xa6, 0xaf, 0x3f, 0xff, 0x63, 0x9c, 0x29, 0x7b, 0x5b,
	0xcf, 0x2e, 0xa5, 0x9e, 0x4f, 0xb2, 0xaa, 0x36, 0x46, 0xd3, 0x9f, 0x89, 0xc1, 0xea, 0x1e, 0xab,
	0xdc, 0x19, 0xee, 0xbf, 0x00, 0x00, 0x00, 0xff, 0xff, 0xc3, 0x92, 0xe3, 0x2f, 0x21, 0x07, 0x00,
	0x00,
}