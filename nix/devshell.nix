{ pkgs, perSystem }:
perSystem.devshell.mkShell {
  packages = [
    # go
    pkgs.air
    pkgs.go
    pkgs.goreleaser
    pkgs.revive

    # k8s
    pkgs.k9s
    pkgs.kind
    pkgs.kubectl
    pkgs.kubelogin-oidc

    # other
    perSystem.self.formatter
    pkgs.just
  ];

  env = [
    {
      name = "NIX_PATH";
      value = "nixpkgs=${toString pkgs.path}";
    }
    {
      name = "NIX_DIR";
      eval = "$PRJ_ROOT/nix";
    }
  ];

  commands = [
    {
      name = "k";
      category = "ops";
      help = "Shorter alias for kubectl";
      command = ''${pkgs.kubectl}/bin/kubectl "$@"'';
    }
    {
      name = "kvs";
      category = "Ops";
      help = "kubectl view-secret alias";
      command = ''${pkgs.kubectl-view-secret}/bin/kubectl-view-secret "$@"'';
    }
    {
      name = "kns";
      category = "ops";
      help = "Switch kubernetes namespaces";
      command = ''kubens "$@"'';
    }
  ];
}
