{
  description = "claat-skill development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          # このリポジトリが必要とするのは Go だけ（claat-tools は標準ライブラリのみ）
          buildInputs = [ pkgs.go ];
        };
      });
}
