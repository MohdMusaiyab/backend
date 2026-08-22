echo "Running Database Migrations..."

for file in $(ls migrations/*.up.sql | sort); do
  echo "Executing $file..."
  sudo docker exec -i notification_db psql -U postgres -d notification_db < "$file"
done

echo "Migrations and Seeding Complete! ✅"
