{
  description = "UFO - Universal Fast Orderer";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    osmosis-patches.url = "path:/Users/hxrts/projects/timewave/ufo/patches/osmosis";
  };

  outputs = { self, nixpkgs, flake-utils, osmosis-patches }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        
        # Import the modular nix files
        packages-module = import ./nix/packages.nix { inherit pkgs; };
        python-module = import ./nix/python.nix { inherit pkgs; };
        scripts-module = import ./nix/scripts.nix { inherit pkgs; inherit self; };
       
        # Extract values from modules 
        ufo-binaries = packages-module.ufo-binaries;
        all-binaries = packages-module.all-binaries;
        ufo-test-runner = packages-module.ufo-test-runner;
        
        python-viz-env = python-module.python-viz-env;
        jupyter-app = python-module.jupyter-app;
        
        build-osmosis-ufo-script = scripts-module.build-osmosis-ufo-script;
        
        # Import devShells module and pass required dependencies
        devShells-module = import ./nix/devShells.nix { 
          inherit pkgs;
          inherit python-viz-env;
          inherit jupyter-app;
          inherit build-osmosis-ufo-script;
        };
        
        # Import apps module and pass required dependencies
        apps-module = import ./nix/apps.nix {
          inherit pkgs;
          inherit self;
          inherit ufo-binaries;
          inherit jupyter-app;
          inherit all-binaries;
          inherit build-osmosis-ufo-script;
        };
        
        # Import test-apps module
        test-apps-module = import ./nix/test-apps.nix {
          inherit pkgs;
          inherit self;
        };
        
      in
      {
        packages = {
          inherit (packages-module) ufo-binaries all-binaries ufo-test-runner;
          default = ufo-binaries;
          
          # Expose Python environment with visualization packages
          python-viz = python-viz-env;
          
          # Expose Jupyter application
          jupyter = jupyter-app;
        };
        
        # Set up dev shells from module
        devShells = devShells-module;
        
        # Expose runnable commands by merging the apps and test-apps modules
        apps = apps-module // test-apps-module;
      }
    );
}
