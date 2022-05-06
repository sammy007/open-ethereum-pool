



if [ ! -f "build/env.bat" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
set workspace="T:\open-ethereum-pool\build\_workspace"
set root="T:\open-ethereum-pool"
set ethdir="T:\open-ethereum-pool\build\_workspace/src/github.com/sammy007"
#if [ ! -L "T:\open-ethereum-pool\build\_workspace\open-ethereum-pool" ]; then
    mkdir -p "$ethdir"
    cd "$ethdir"
    ln -s ../../../../../. open-ethereum-pool
    cd "T:\open-ethereum-pool"
#fi

# Set up the environment to use the workspace.
# Also add Godeps workspace so we build using canned dependencies.
GOPATH="$workspace"
GOBIN="$PWD/build/bin"
export GOPATH GOBIN

# Run the command inside the workspace.
cd "$ethdir/open-ethereum-pool"
PWD="$ethdir/open-ethereum-pool"

# Launch the arguments with the configured environment.
exec "$@"