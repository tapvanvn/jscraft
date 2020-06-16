DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

mkdir -p "$DIR/.tmp"
mkdir -p "$DIR/.tmp/layout_1"

pushd $DIR/../src 
    CGO_ENABLED=0 go build -o $DIR/.tmp/jscraft
    if [ $? -ne 0 ]; then 
        echo "build fail"
        exit 1
    fi
popd

$DIR/.tmp/jscraft $DIR/template $DIR/layout_1 $DIR/.tmp/layout_1

