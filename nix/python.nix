# python.nix - Python environment and Jupyter app
{ pkgs }:

let
  # Python environment with visualization packages
  python-viz-env = pkgs.python3.withPackages (ps: with ps; [
    pandas
    matplotlib
    numpy
    ipython
    jupyter
    notebook
    jupyterlab
    ipykernel
    ipywidgets
    jupyter_client
    nbformat
    nbconvert
    seaborn
  ]);
  
  # Create a Jupyter application that includes the right environment
  jupyter-app = pkgs.writeShellScriptBin "ufo-jupyter" ''
    export JUPYTER_PATH=${python-viz-env}/share/jupyter
    export JUPYTER_CONFIG_DIR=''${JUPYTER_CONFIG_DIR:-~/.jupyter}
    
    # Set up Jupyter kernel if it doesn't exist
    if ! [ -d "$JUPYTER_CONFIG_DIR/kernels/ufo-benchmark" ]; then
      echo "Setting up Jupyter kernel for UFO benchmark notebooks..."
      ${python-viz-env}/bin/python -m ipykernel install --user --name ufo-benchmark --display-name "UFO Benchmark"
    fi
    
    echo "Starting Jupyter with the UFO benchmark environment..."
    ${python-viz-env}/bin/jupyter "$@"
  '';
in
{
  inherit python-viz-env;
  inherit jupyter-app;
} 