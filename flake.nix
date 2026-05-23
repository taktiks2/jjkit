{
  description = "Go devShell (default: pkgs.go = latest stable)";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
    let
      forAllSystems = nixpkgs.lib.genAttrs [
        "aarch64-darwin"
        "x86_64-darwin"
        "aarch64-linux"
        "x86_64-linux"
      ];
    in
    {
      devShells = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            name = "go-shell";

            packages = with pkgs; [
              go
              gopls
              gotools
              delve
              golangci-lint
              just
              jq
            ];

            shellHook = ''
              export GOBIN="$PWD/.gobin"
              export PATH="$GOBIN:$PATH"
              mkdir -p "$GOBIN"
              echo "→ devShell: go ($(go version | awk '{print $3}'))"
            '';
          };
        });
    };
}
