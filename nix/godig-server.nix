{ lib, buildGoModule }:

buildGoModule rec {
  pname = "godig-server";
  version = "0.2.0";

  src = ../.;

  subPackages = [ "cmd/server" ];

  vendorHash = "sha256-rG5eCE2si6GRboO+Skuc+1O5lxkjDNte8eLCG4vcFJs=";

  meta = with lib; {
    description = "Godig tunnel server - accepts service connections and routes HTTP requests";
    homepage = "https://github.com/AYM1607/godig";
    license = licenses.mit;
    maintainers = [ ];
  };
}
