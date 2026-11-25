package ldap

import (
	"testing"

	"github.com/stretchr/testify/require"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

// requireLDAPResult asserts RFC 4511 LDAPResult:
// SEQUENCE { resultCode ENUMERATED, matchedDN OCTET STRING, diagnosticMessage OCTET STRING }.
// The packet is encoded to bytes and decoded again, as a third-party client sees it.
func requireLDAPResult(t *testing.T, body *ber.Packet, code int64, matchedDn, message string) {
	t.Helper()

	decoded, err := ber.DecodePacketErr(body.Bytes())
	require.NoError(t, err)
	require.Len(t, decoded.Children, 3, "LDAPResult needs resultCode, matchedDN and diagnosticMessage")

	require.EqualValues(t, ber.TagEnumerated, decoded.Children[0].Tag)
	require.Equal(t, code, decoded.Children[0].Value)

	require.EqualValues(t, ber.TagOctetString, decoded.Children[1].Tag)
	require.Equal(t, matchedDn, decoded.Children[1].Value)

	require.EqualValues(t, ber.TagOctetString, decoded.Children[2].Tag)
	require.Equal(t, message, decoded.Children[2].Value)
}

func newResponseEnvelope() *ber.Packet {
	return ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Response")
}

func TestCompareResponse_encode_isLDAPResult(t *testing.T) {
	envelope := newResponseEnvelope()
	(&CompareResponse{ResultCode: CompareTrue, MatchedDn: "dc=example", Message: "foo"}).encode(envelope)

	require.Len(t, envelope.Children, 1)
	requireLDAPResult(t, envelope.Children[0], int64(CompareTrue), "dc=example", "foo")
}

func TestModifyDNResponse_encode_isLDAPResult(t *testing.T) {
	envelope := newResponseEnvelope()
	(&ModifyDNResponse{ResultCode: Success, MatchedDn: "dc=example", Message: "bar"}).encode(envelope)

	require.Len(t, envelope.Children, 1)
	requireLDAPResult(t, envelope.Children[0], int64(Success), "dc=example", "bar")
}

func TestModifyDNResponse_encode_decodedByClient(t *testing.T) {
	envelope := newResponseEnvelope()
	(&ModifyDNResponse{ResultCode: Success, MatchedDn: "dc=example", Message: "bar"}).encode(envelope)

	decoded, err := ber.DecodePacketErr(envelope.Children[0].Bytes())
	require.NoError(t, err)
	r, err := decodeModifyDnResponse(decoded)
	require.NoError(t, err)
	require.Equal(t, "dc=example", r.MatchedDn)
	require.Equal(t, "bar", r.Message)
}

func TestCompareResponse_encode_decodedByClient(t *testing.T) {
	envelope := newResponseEnvelope()
	(&CompareResponse{ResultCode: CompareFalse, MatchedDn: "dc=example", Message: "foo"}).encode(envelope)

	decoded, err := ber.DecodePacketErr(envelope.Children[0].Bytes())
	require.NoError(t, err)
	r, err := decodeCompareResponse(decoded)
	require.NoError(t, err)
	require.Equal(t, CompareFalse, r.ResultCode)
	require.Equal(t, "dc=example", r.MatchedDn)
	require.Equal(t, "foo", r.Message)
}
