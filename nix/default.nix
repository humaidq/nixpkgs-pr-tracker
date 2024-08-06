{
  buildGoModule,
  git,
  just,
  pnpm,
  nodejs,
  lib,
  stdenv,
}: let
  version = "0.1-beta";
  webapp = stdenv.mkDerivation (finalAttrs: {
    src = ./../app;
    pname = "nixpkgs-pr-tracker-webapp";
    name = finalAttrs.pname;
    inherit version;
    pnpmDeps = pnpm.fetchDeps {
      inherit (finalAttrs) pname version src;
      hash = "sha256-XQ1gDjUtVCfXtg3KWdsad5gMui4th9Zq5mPX3soY6uc=";
    };
    nativeBuildInputs = [
      nodejs
      pnpm.configHook
    ];
    postBuild = ''
      pnpm run build
    '';
    installPhase = ''
      cp -r dist/ $out
    '';
  });
in
  buildGoModule {
    name = "nixpkgs-pr-tracker";
    inherit version;
    src = ./..;

    nativeBuildInputs = [
      git
      just
      nodejs
    ];

    inherit webapp;

    prePatch = ''
      cp -r ${webapp} app/dist
      rm app/public/embed.go
    '';
    vendorHash = "sha256-t5qgehVAe7fRSZ38l8SXfwUexrUkdOV9854x7YdVxco=";

    meta = with lib; {
      description = "Nixpkgs PR Tracker";
      homepage = "https://github.com/humaidq/nix-tracker";
      license = licenses.agpl3Plus;
    };
  }
