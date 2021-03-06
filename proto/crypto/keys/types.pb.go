// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/crypto/keys/types.proto

package keys

import (
	bytes "bytes"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// PublicKey defines the keys available for use with Tendermint Validators
type PublicKey struct {
	// Types that are valid to be assigned to Sum:
	//	*PublicKey_Ed25519
	Sum isPublicKey_Sum `protobuf_oneof:"sum"`
}

func (m *PublicKey) Reset()         { *m = PublicKey{} }
func (m *PublicKey) String() string { return proto.CompactTextString(m) }
func (*PublicKey) ProtoMessage()    {}
func (*PublicKey) Descriptor() ([]byte, []int) {
	return fileDescriptor_943d79b57ec0188f, []int{0}
}
func (m *PublicKey) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PublicKey) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PublicKey.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PublicKey) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PublicKey.Merge(m, src)
}
func (m *PublicKey) XXX_Size() int {
	return m.Size()
}
func (m *PublicKey) XXX_DiscardUnknown() {
	xxx_messageInfo_PublicKey.DiscardUnknown(m)
}

var xxx_messageInfo_PublicKey proto.InternalMessageInfo

type isPublicKey_Sum interface {
	isPublicKey_Sum()
	Equal(interface{}) bool
	MarshalTo([]byte) (int, error)
	Size() int
	Compare(interface{}) int
}

type PublicKey_Ed25519 struct {
	Ed25519 []byte `protobuf:"bytes,1,opt,name=ed25519,proto3,oneof" json:"ed25519,omitempty"`
}

func (*PublicKey_Ed25519) isPublicKey_Sum() {}

func (m *PublicKey) GetSum() isPublicKey_Sum {
	if m != nil {
		return m.Sum
	}
	return nil
}

func (m *PublicKey) GetEd25519() []byte {
	if x, ok := m.GetSum().(*PublicKey_Ed25519); ok {
		return x.Ed25519
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*PublicKey) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*PublicKey_Ed25519)(nil),
	}
}

// PrivateKey defines the keys available for use with Tendermint Validators
// WARNING PrivateKey is used for internal purposes only
type PrivateKey struct {
	// Types that are valid to be assigned to Sum:
	//	*PrivateKey_Ed25519
	Sum isPrivateKey_Sum `protobuf_oneof:"sum"`
}

func (m *PrivateKey) Reset()         { *m = PrivateKey{} }
func (m *PrivateKey) String() string { return proto.CompactTextString(m) }
func (*PrivateKey) ProtoMessage()    {}
func (*PrivateKey) Descriptor() ([]byte, []int) {
	return fileDescriptor_943d79b57ec0188f, []int{1}
}
func (m *PrivateKey) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PrivateKey) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PrivateKey.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PrivateKey) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PrivateKey.Merge(m, src)
}
func (m *PrivateKey) XXX_Size() int {
	return m.Size()
}
func (m *PrivateKey) XXX_DiscardUnknown() {
	xxx_messageInfo_PrivateKey.DiscardUnknown(m)
}

var xxx_messageInfo_PrivateKey proto.InternalMessageInfo

type isPrivateKey_Sum interface {
	isPrivateKey_Sum()
	MarshalTo([]byte) (int, error)
	Size() int
}

type PrivateKey_Ed25519 struct {
	Ed25519 []byte `protobuf:"bytes,1,opt,name=ed25519,proto3,oneof" json:"ed25519,omitempty"`
}

func (*PrivateKey_Ed25519) isPrivateKey_Sum() {}

func (m *PrivateKey) GetSum() isPrivateKey_Sum {
	if m != nil {
		return m.Sum
	}
	return nil
}

func (m *PrivateKey) GetEd25519() []byte {
	if x, ok := m.GetSum().(*PrivateKey_Ed25519); ok {
		return x.Ed25519
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*PrivateKey) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*PrivateKey_Ed25519)(nil),
	}
}

func init() {
	proto.RegisterType((*PublicKey)(nil), "tendermint.crypto.keys.PublicKey")
	proto.RegisterType((*PrivateKey)(nil), "tendermint.crypto.keys.PrivateKey")
}

func init() { proto.RegisterFile("proto/crypto/keys/types.proto", fileDescriptor_943d79b57ec0188f) }

var fileDescriptor_943d79b57ec0188f = []byte{
	// 213 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x2d, 0x28, 0xca, 0x2f,
	0xc9, 0xd7, 0x4f, 0x2e, 0xaa, 0x2c, 0x28, 0xc9, 0xd7, 0xcf, 0x4e, 0xad, 0x2c, 0xd6, 0x2f, 0xa9,
	0x2c, 0x48, 0x2d, 0xd6, 0x03, 0x8b, 0x0b, 0x89, 0x95, 0xa4, 0xe6, 0xa5, 0xa4, 0x16, 0xe5, 0x66,
	0xe6, 0x95, 0xe8, 0x41, 0xd4, 0xe8, 0x81, 0xd4, 0x48, 0xa9, 0x95, 0x64, 0x64, 0x16, 0xa5, 0xc4,
	0x17, 0x24, 0x16, 0x95, 0x54, 0xea, 0x43, 0x8c, 0x48, 0xcf, 0x4f, 0xcf, 0x47, 0xb0, 0x20, 0xfa,
	0x95, 0x2c, 0xb8, 0x38, 0x03, 0x4a, 0x93, 0x72, 0x32, 0x93, 0xbd, 0x53, 0x2b, 0x85, 0xa4, 0xb8,
	0xd8, 0x53, 0x53, 0x8c, 0x4c, 0x4d, 0x0d, 0x2d, 0x25, 0x18, 0x15, 0x18, 0x35, 0x78, 0x3c, 0x18,
	0x82, 0x60, 0x02, 0x56, 0x1c, 0x2f, 0x16, 0xc8, 0x33, 0xbe, 0x58, 0x28, 0xcf, 0xe8, 0xc4, 0xca,
	0xc5, 0x5c, 0x5c, 0x9a, 0xab, 0xa4, 0xcf, 0xc5, 0x15, 0x50, 0x94, 0x59, 0x96, 0x58, 0x92, 0x4a,
	0x40, 0x2b, 0x54, 0x83, 0x93, 0xcf, 0x89, 0x47, 0x72, 0x8c, 0x17, 0x1e, 0xc9, 0x31, 0x3e, 0x78,
	0x24, 0xc7, 0x38, 0xe1, 0xb1, 0x1c, 0xc3, 0x85, 0xc7, 0x72, 0x0c, 0x37, 0x1e, 0xcb, 0x31, 0x44,
	0x19, 0xa5, 0x67, 0x96, 0x64, 0x94, 0x26, 0xe9, 0x25, 0xe7, 0xe7, 0xea, 0x23, 0xfc, 0x83, 0xcc,
	0xc4, 0x08, 0x84, 0x24, 0x36, 0xb0, 0x90, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0x65, 0xc0, 0xc2,
	0x6d, 0x20, 0x01, 0x00, 0x00,
}

func (this *PublicKey) Compare(that interface{}) int {
	if that == nil {
		if this == nil {
			return 0
		}
		return 1
	}

	that1, ok := that.(*PublicKey)
	if !ok {
		that2, ok := that.(PublicKey)
		if ok {
			that1 = &that2
		} else {
			return 1
		}
	}
	if that1 == nil {
		if this == nil {
			return 0
		}
		return 1
	} else if this == nil {
		return -1
	}
	if that1.Sum == nil {
		if this.Sum != nil {
			return 1
		}
	} else if this.Sum == nil {
		return -1
	} else {
		thisType := -1
		switch this.Sum.(type) {
		case *PublicKey_Ed25519:
			thisType = 0
		default:
			panic(fmt.Sprintf("compare: unexpected type %T in oneof", this.Sum))
		}
		that1Type := -1
		switch that1.Sum.(type) {
		case *PublicKey_Ed25519:
			that1Type = 0
		default:
			panic(fmt.Sprintf("compare: unexpected type %T in oneof", that1.Sum))
		}
		if thisType == that1Type {
			if c := this.Sum.Compare(that1.Sum); c != 0 {
				return c
			}
		} else if thisType < that1Type {
			return -1
		} else if thisType > that1Type {
			return 1
		}
	}
	return 0
}
func (this *PublicKey_Ed25519) Compare(that interface{}) int {
	if that == nil {
		if this == nil {
			return 0
		}
		return 1
	}

	that1, ok := that.(*PublicKey_Ed25519)
	if !ok {
		that2, ok := that.(PublicKey_Ed25519)
		if ok {
			that1 = &that2
		} else {
			return 1
		}
	}
	if that1 == nil {
		if this == nil {
			return 0
		}
		return 1
	} else if this == nil {
		return -1
	}
	if c := bytes.Compare(this.Ed25519, that1.Ed25519); c != 0 {
		return c
	}
	return 0
}
func (this *PublicKey) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*PublicKey)
	if !ok {
		that2, ok := that.(PublicKey)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if that1.Sum == nil {
		if this.Sum != nil {
			return false
		}
	} else if this.Sum == nil {
		return false
	} else if !this.Sum.Equal(that1.Sum) {
		return false
	}
	return true
}
func (this *PublicKey_Ed25519) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*PublicKey_Ed25519)
	if !ok {
		that2, ok := that.(PublicKey_Ed25519)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if !bytes.Equal(this.Ed25519, that1.Ed25519) {
		return false
	}
	return true
}
func (m *PublicKey) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PublicKey) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PublicKey) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Sum != nil {
		{
			size := m.Sum.Size()
			i -= size
			if _, err := m.Sum.MarshalTo(dAtA[i:]); err != nil {
				return 0, err
			}
		}
	}
	return len(dAtA) - i, nil
}

func (m *PublicKey_Ed25519) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PublicKey_Ed25519) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.Ed25519 != nil {
		i -= len(m.Ed25519)
		copy(dAtA[i:], m.Ed25519)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Ed25519)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}
func (m *PrivateKey) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PrivateKey) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PrivateKey) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Sum != nil {
		{
			size := m.Sum.Size()
			i -= size
			if _, err := m.Sum.MarshalTo(dAtA[i:]); err != nil {
				return 0, err
			}
		}
	}
	return len(dAtA) - i, nil
}

func (m *PrivateKey_Ed25519) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PrivateKey_Ed25519) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.Ed25519 != nil {
		i -= len(m.Ed25519)
		copy(dAtA[i:], m.Ed25519)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Ed25519)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}
func encodeVarintTypes(dAtA []byte, offset int, v uint64) int {
	offset -= sovTypes(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *PublicKey) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Sum != nil {
		n += m.Sum.Size()
	}
	return n
}

func (m *PublicKey_Ed25519) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Ed25519 != nil {
		l = len(m.Ed25519)
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}
func (m *PrivateKey) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Sum != nil {
		n += m.Sum.Size()
	}
	return n
}

func (m *PrivateKey_Ed25519) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Ed25519 != nil {
		l = len(m.Ed25519)
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}

func sovTypes(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTypes(x uint64) (n int) {
	return sovTypes(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *PublicKey) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PublicKey: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PublicKey: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ed25519", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := make([]byte, postIndex-iNdEx)
			copy(v, dAtA[iNdEx:postIndex])
			m.Sum = &PublicKey_Ed25519{v}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *PrivateKey) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PrivateKey: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PrivateKey: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ed25519", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := make([]byte, postIndex-iNdEx)
			copy(v, dAtA[iNdEx:postIndex])
			m.Sum = &PrivateKey_Ed25519{v}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTypes(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthTypes
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTypes
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTypes
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTypes        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTypes          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTypes = fmt.Errorf("proto: unexpected end of group")
)
