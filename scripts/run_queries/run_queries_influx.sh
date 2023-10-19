#!/bin/bash

# Ensure runner is available
EXE_FILE_NAME=${EXE_FILE_NAME:-$(which tsbs_run_queries_influx)}
if [[ -z "$EXE_FILE_NAME" ]]; then
    echo "tsbs_run_queries_influx not available. It is not specified explicitly and not found in \$PATH"
    exit 1
fi

# Default queries folder
BULK_DATA_DIR=${BULK_DATA_DIR:-"/tmp/bulk_queries"}
DATABASE_NAME=${DATABASE_NAME:-"benchmark"}
MAX_QUERIES=${MAX_QUERIES:-"0"}
# How many concurrent worker would run queries - match num of cores, or default to 4
NUM_WORKERS=${NUM_WORKERS:-$(grep -c ^processor /proc/cpuinfo 2> /dev/null || echo 4)}
QUERIES_PRINT_INTERVAL=100
DEBUG=0

DATABASE_PORT=${DATABASE_PORT:-8086}
INFLUX_AUTH_TOKEN=${$INFLUX_AUTH_TOKEN:-""}

EXE_DIR=${EXE_DIR:-$(dirname $0)}


until curl http://${DATABASE_HOST}:${DATABASE_PORT}/ping 2>/dev/null; do
    echo "Waiting for InfluxDB"
    sleep 1
done

#
# Run test for one file
#
function run_file() {
    # $FULL_DATA_FILE_NAME:  /full/path/to/file_with.ext
    # $DATA_FILE_NAME:       file_with.ext
    # $DIR:                  /full/path/to
    # $EXTENSION:            ext
    # NO_EXT_DATA_FILE_NAME: file_with
    FULL_DATA_FILE_NAME=$1

    RESULTS_DIR=$(dirname "${FULL_DATA_FILE_NAME}")

    DATA_FILE_NAME=$(basename -- "${FULL_DATA_FILE_NAME}")

#    echo $DATA_FILE_NAME

    DIR=$(dirname "${FULL_DATA_FILE_NAME}")
    EXTENSION="${DATA_FILE_NAME##*.}"
    NO_EXT_DATA_FILE_NAME="${DATA_FILE_NAME%.*}"

    if [ "${EXTENSION}" == "gz" ]; then
        GUNZIP="gunzip"
    else
        GUNZIP="cat"
    fi

    # Several options on how to name results file
    #OUT_FULL_FILE_NAME="${DIR}/result_${DATA_FILE_NAME}"
    OUT_FULL_FILE_NAME="${RESULTS_DIR}/result_${NO_EXT_DATA_FILE_NAME}_${run}.out"
    #OUT_FULL_FILE_NAME="${DIR}/${NO_EXT_DATA_FILE_NAME}.out"
    HDR_FULL_FILE_NAME="${RESULTS_DIR}/HDR_TXT_result_${NO_EXT_DATA_FILE_NAME}_${run}.out"

    echo "Running ${DATA_FILE_NAME}"
    echo "    Saving results to ${OUT_FULL_FILE_NAME}"
    echo "    Saving HDR results to ${HDR_FULL_FILE_NAME}"

    cat $FULL_DATA_FILE_NAME |
        $GUNZIP |
        $EXE_FILE_NAME \
            --max-queries=${MAX_QUERIES} \
            --db-name=${DATABASE_NAME} \
            --workers=${NUM_WORKERS} \
            --print-interval=${QUERIES_PRINT_INTERVAL} \
            --hdr-latencies=${HDR_FULL_FILE_NAME} \
            --prewarm-queries \
            --auth-token $INFLUX_AUTH_TOKEN \
            --debug=${DEBUG} \
            --urls=http://${DATABASE_HOST}:${DATABASE_PORT} |
        tee $OUT_FULL_FILE_NAME

}

if [ "$#" -gt 0 ]; then
    echo "Have $# files specified as params"
    for FULL_DATA_FILE_NAME in "$@"; do
        run_file $FULL_DATA_FILE_NAME
    done
else
    echo "Do not have any files specified - run from default queries folder as ${BULK_DATA_DIR}/queries_queries*"
    for FULL_DATA_FILE_NAME in "${BULK_DATA_DIR}/queries_influx"*; do
        run_file $FULL_DATA_FILE_NAME
    done
fi
