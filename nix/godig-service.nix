{ lib, buildGoModule }:

buildGoModule rec {
  pname = "godig-service";
  version = "0.2.0";

  src = ../.;

  subPackages = [ "cmd/service" ];

  vendorHash = "sha256-rG5eCE2si6GRboO+Skuc+1O5lxkjDNte8eLCG4vcFJs=";

  meta = with lib; {
    description = "Godig tunnel client - connects to server and exposes local services";
    homepage = "https://github.com/AYM1607/godig";
    license = licenses.mit;
    maintainers = [ ];
  };
}
