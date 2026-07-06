# Ad Analytics & Management Portal (High-Throughput Tooling Demo)

This full-stack system is designed to simulate live production workloads for modern ad-tech infrastructures. Built with an eye for performance optimization, high concurrency handling, and clean database structures.

### 🚀 Stack Specifications
- **Backend:** Go (Golang) featuring standard concurrency paradigms and multiplexed memory router pipelines.
- **Frontend:** React.js structured inside an optimized modular environment managed via Vite and styled dynamically with utility classes.
- **Database Engine:** PostgreSQL managing core business entities (relational campaigns) layered next to structural data design built for high analytical throughput tracking.
- **MCP Core Spec Alignment:** Emulates basic endpoint frameworks aligned to LLM server querying schemas.

### ⚡ Quick Start Running Locally
1. Run a local PostgreSQL instance and update `DATABASE_URL` or ensure default credentials (`postgres/postgres`) map to a DB named `adtech_db`.
2. **Launch Engine Server:**
   ```bash
   cd backend
   go run .
