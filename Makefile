build:
	cd frontend && npm run build
	mv frontend/build backend/
run:
	cd backend && go run .