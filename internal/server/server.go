package server

type Server struct {
}

func New() *Server {
	return &Server{}
}

func (s *Server) Start() error {
	return nil
}

func (s *Server) Stop() error {
	return nil
}
