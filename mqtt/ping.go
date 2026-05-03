package mqtt

type PingRequest struct{}

type PingResponse struct{}

func (r *PingRequest) Read(_ *Decoder, _ *Header) {}

func (r *PingRequest) Write(_ *Encoder, _ *Header) {}

func (r *PingResponse) Read(_ *Decoder, _ *Header) {}

func (r *PingResponse) Write(_ *Encoder, _ *Header) {}
