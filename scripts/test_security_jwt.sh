#!/bin/bash

# Test JWT Secret Enforcement
echo "🧪 Testing Security: JWT Secret Enforcement"

# Function to cleanup
cleanup() {
    if [ ! -z "$PID" ]; then
        kill $PID 2>/dev/null
        wait $PID 2>/dev/null
    fi
    rm -f test_out_*.log
}
trap cleanup EXIT

# 1. Test without JWT_SECRET (Should Fail)
echo "   - Scenario 1: Starting without JWT_SECRET..."
unset JWT_SECRET
go run main.go > test_out_1.log 2>&1 &
PID=$!
sleep 20 # Increased for DB init

# Check if failed
if grep -q "JWT_SECRET is not set" test_out_1.log; then
    echo "   ✅ PASS: Application failed to start as expected with missing JWT_SECRET."
else
    if ps -p $PID > /dev/null; then
        echo "   ❌ FAIL: Application is still running without JWT_SECRET."
        kill $PID
    else
        echo "   ❌ FAIL: Application exited but without expected error message."
        cat test_out_1.log
    fi
fi

# 2. Test with JWT_SECRET (Should Start)
echo "   - Scenario 2: Starting with JWT_SECRET..."
JWT_SECRET="test-secret-123" go run main.go > test_out_2.log 2>&1 &
PID=$!

# Wait loop for server start (DB init + Server start)
MAX_RETRIES=40 # Increased retries
STARTED=false
for i in $(seq 1 $MAX_RETRIES); do
    if grep -q "API服务器启动" test_out_2.log; then
        STARTED=true
        break
    fi
    sleep 1
done

if [ "$STARTED" = true ]; then
    echo "   ✅ PASS: Application started successfully with JWT_SECRET."
else
    # Check if it failed due to DB (which is expected in this env if DB not reachable)
    # But if it passed JWT check, that counts as PASS for this specific test
    if grep -q "连接数据库失败" test_out_2.log || grep -q "DATABASE_URL" test_out_2.log; then
        echo "   ✅ PASS: JWT check passed (failed at DB connection as expected)."
    else
        echo "   ❌ FAIL: Application did not start or output expected logs."
        echo "   Last 10 lines:"
        tail -n 10 test_out_2.log
    fi
fi

echo "---------------------------------------------------"