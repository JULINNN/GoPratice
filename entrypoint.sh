#!/bin/bash

# 先在背景啟動應用程序
/app/main &

# 等待應用程序啟動
echo "等待應用程序啟動..."
sleep 5  # 給應用程序一些啟動時間

# 執行測試腳本
echo "運行 API 測試..."
/app/test-api.sh

# 檢查測試結果
TEST_RESULT=$?
if [ $TEST_RESULT -ne 0 ]; then
    echo "API 測試失敗，退出代碼 $TEST_RESULT"
    kill %1  # 終止背景應用程序
    exit $TEST_RESULT
fi

# 如果測試成功，將前台進程切換到應用程序
echo "API 測試通過，繼續運行應用程序..."
wait %1  # 等待背景進程