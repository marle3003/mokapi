package protocol

//type GroupMetadata struct {
//	Version  int16
//	Topics   []string
//	UserData []byte
//}
//
//func ParseGroupMetadata(b []byte) {
//	r := bytes.NewReader(b)
//	d := NewDecoder(r, len(b))
//	g := GroupMetadata{}
//	g.Version = d.readInt16()
//	d.decodeArray(reflect.ValueOf(g.Topics), (*Decoder).decodeString)
//	g.UserData = d.readBytes()
//}
//
//func (g GroupMetadata) Write(w io.Writer) {
//	e := NewEncoder(w)
//	e.writeInt16(g.Version)
//	e.encodeArray(reflect.ValueOf(g.Topics), (*Encoder).encodeString)
//	e.writeBytes(g.UserData)
//}
