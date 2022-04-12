package egorm

import (
	"testing"
)

func Test_peerInfo(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name         string
		args         args
		wantHostname string
		wantPort     int
	}{
		{
			name: "testIpv4",
			args: args{
				addr: "127.0.0.1:3306",
			},
			wantHostname: "127.0.0.1",
			wantPort:     3306,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHostname, gotPort := peerInfo(tt.args.addr)
			if gotHostname != tt.wantHostname {
				t.Errorf("peerInfo() gotHostname = %v, want %v", gotHostname, tt.wantHostname)
			}
			if gotPort != tt.wantPort {
				t.Errorf("peerInfo() gotPort = %v, want %v", gotPort, tt.wantPort)
			}
		})
	}
}
