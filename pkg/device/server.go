package device

func (s *server) newPacket() *Packet {
	return &Packet{
		server: s,
	}
}
