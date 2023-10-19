# tsbs - influx v2 性能测试示例

# 准备环境
- 创建测试用buket
```shell
influx bucket create -n bucket-perf -o ymatrix -r 0 -t qWUdCcPDbCY1X89P9_5g_fwIKHMFBoLtLLc1aWLk79ydqyBi8HIlLDPkivK4zN_dlw2OqXm7OXXKxMQXxNq2hQ==
```
- 创建一个 DBRP mapping 兼容 InfluxDB 1.x compatibility API (official docs).
```shell
influx v1 dbrp create --db benchmark --org ymatrix --token qWUdCcPDbCY1X89P9_5g_fwIKHMFBoLtLLc1aWLk79ydqyBi8HIlLDPkivK4zN_dlw2OqXm7OXXKxMQXxNq2hQ== --rp 0 --bucket-id `influx bucket ls --name bucket-perf --org ymatrix --token qWUdCcPDbCY1X89P9_5g_fwIKHMFBoLtLLc1aWLk79ydqyBi8HIlLDPkivK4zN_dlw2OqXm7OXXKxMQXxNq2hQ==| awk -v i=2 -v j=1 'FNR == i {print $j}'` --default 
```
# 编译
```shell
git clone https://github.com/Cousin-Zan/tsbs
```
```shell
cd tsbs
go install ./...
```
# 生成数据
```shell
FORMATS="influx" SCALE=20 SEED=123 USE_CASE="devops"\
    TS_START="2016-01-01T00:00:00Z" \
    TS_END="2016-01-04T00:00:01Z" \
    BULK_DATA_DIR="/tmp/bulk_data" scripts/generate_data.sh
```
# 生成查询
```shell
FORMATS="influx" SCALE=4000 SEED=123 USE_CASE="devops"\
    TS_START="2016-01-01T00:00:00Z" \
    TS_END="2016-01-04T00:00:01Z" \
    BULK_DATA_DIR="/tmp/bulk_queries" scripts/generate_queries.sh
```
# 加载数据
```shell
INFLUX_AUTH_TOKEN="qWUdCcPDbCY1X89P9_5g_fwIKHMFBoLtLLc1aWLk79ydqyBi8HIlLDPkivK4zN_dlw2OqXm7OXXKxMQXxNq2hQ==" DATABASE_HOST="82.157.59.105" DATABASE_PORT="8086" DATABASE_NAME="benchmark" scripts/load/load_influx.sh
```
# 执行查询
```shell
INFLUX_AUTH_TOKEN="qWUdCcPDbCY1X89P9_5g_fwIKHMFBoLtLLc1aWLk79ydqyBi8HIlLDPkivK4zN_dlw2OqXm7OXXKxMQXxNq2hQ==" DATABASE_HOST="82.157.59.105" DATABASE_PORT="8086" BULK_DATA_DIR="/tmp/bulk_queries" DATABASE_NAME="benchmark" MAX_QUERIES=10 NUM_WORKERS=1 scripts/run_queries/run_queries_influx.sh
```