{
  pkgs,
  buildGoApplication,
}: let
  pname = "jwt-tui";
  version = "0.1.0";
  ldflags = ["-s" "-w" "-X main.version=${version}"];
  pkg = buildGoApplication {
    inherit pname version ldflags;
    src = ../.;
    modules = ./gomod2nix.toml;
    meta = with pkgs.lib; {
      description = "A TUI for inspecting, editing, and signing JSON Web Tokens (JWTs).";
      homepage = "https://github.com/anotherhadi/jwt-tui";
      platforms = platforms.unix;
    };
  };
in {
  "${pname}" = pkg;
  default = pkg;
}
