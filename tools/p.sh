# دالة لحذف أي عملية تعمل على بورت معين
p() {
  if [ -z "$1" ]; then
    echo "Usage: p [port_number] (e.g., p 8080)"
    return 1
  fi
  
  # البحث عن الـ PID وقتله فوراً
  PID=$(sudo lsof -t -i:$1)
  
  if [ -z "$PID" ]; then
    echo "No process found on port $1"
  else
    echo "Killing process $PID on port $1..."
    sudo kill -9 $PID
    echo "Done!"
  fi
}