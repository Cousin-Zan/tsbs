#!/bin/bash

EXE_FILE_NAME=./bin/tsbs_load_redistimeseries

set -x

EXE_DIR=${EXE_DIR:-$(dirname $0)}
source ${EXE_DIR}/query_common.sh
source ${EXE_DIR}/redistimeseries_common.sh

# Ensure RESULTS DIR available
mkdir -p ${RESULTS_DIR}

OUT_FULL_FILE_NAME="${RESULTS_DIR}/${PREFIX}_load_result.out"
echo "Using only 1 worker"
echo "Saving results to ${OUT_FULL_FILE_NAME}"

# Load new data
$EXE_FILE_NAME \
  --file ${DATA_FILE_NAME} \
  --workers=1 \
  --batch-size=${BATCH_SIZE} \
  --reporting-period=${REPORTING_PERIOD} \
  --host=${DATABASE_HOST}:${DATABASE_PORT} ${CLUSTER_FLAG} \
  --connections=${CONNECTIONS} --pipeline=${PIPELINE} |
  tee ${OUT_FULL_FILE_NAME}
