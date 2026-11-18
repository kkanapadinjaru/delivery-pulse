# SAM Assistant Architecture

## 1. Overall component architecture
This is the main diagram. It shows the major components, their responsibilities, and how information flows through the system.

flowchart TB
    subgraph SOURCES["1. Knowledge Sources"]
        LEGACY["Legacy Repository<br/>Classic ASP · SQL Server"]
        MODERN["Modern Repository<br/>.NET · gRPC · React · Helm"]
        DOCS["Product Documentation<br/>Markdown · PDFs · Indentures"]
        ARCH["Architecture Knowledge<br/>ADRs · Design Docs · Engineering Guidelines"]
    end

    subgraph INGESTION["2. Knowledge Ingestion"]
        GIT["Git Synchronization"]
        DISCOVERY["File Discovery<br/>& Change Detection"]
        PARSERS["Source Parsers<br/>SQL · ASP · C# · Markdown · PDF"]
        STRUCT["Structural Extraction<br/>Symbols · Procedures · Dependencies"]
        CHUNK["Semantic Chunking"]
        ENRICH["LLM Semantic Enrichment<br/>Explanations · Concepts · Candidate Relationships"]
        EMBED["Embedding Generation"]
    end

    subgraph STORAGE["3. Knowledge Storage — PostgreSQL"]
        DOCSTORE[("Sources & Documents<br/>Repository · Commit · File Hash")]
        CHUNKSTORE[("Chunks & Symbols<br/>Content · Line Numbers · Metadata")]
        VECTOR[("Vector Index<br/>pgvector")]
        FULLTEXT[("Full-text Index<br/>PostgreSQL FTS")]
        RELATIONS[("Relationships<br/>Calls · Reads · Writes · References")]
        RUNS[("Ingestion Runs<br/>Status · Commit · Timestamps")]
    end

    subgraph KNOWLEDGE["4. Knowledge Application — Python"]
        API["FastAPI"]
        RETRIEVAL["Retrieval Layer"]
        SEMANTIC["Semantic Search"]
        KEYWORD["Keyword / Exact Search"]
        RELATION["Relationship Expansion"]
        HYBRID["Hybrid Ranking & Reranking"]
        RAG["RAG / Assistant Service"]
        PROMPT["Prompt Construction"]
        LLM["LLM Provider Interface"]
        RESPONSE["Grounded Answer<br/>with Source Citations"]
    end

    subgraph CONSUMERS["5. Developer Interfaces"]
        CLI["CLI"]
        CHAT["Internal Chat UI<br/>(React — Later)"]
        MCP["MCP Server"]
        CLAUDE["Claude Code"]
    end

    LEGACY --> GIT
    MODERN --> GIT
    DOCS --> DISCOVERY
    ARCH --> DISCOVERY
    GIT --> DISCOVERY
    DISCOVERY --> PARSERS
    PARSERS --> STRUCT
    STRUCT --> CHUNK
    CHUNK --> ENRICH
    ENRICH --> EMBED

    DISCOVERY --> DOCSTORE
    STRUCT --> CHUNKSTORE
    CHUNK --> CHUNKSTORE
    EMBED --> VECTOR
    CHUNK --> FULLTEXT
    STRUCT --> RELATIONS
    ENRICH --> RELATIONS
    GIT --> RUNS

    CLI --> API
    CHAT --> API
    MCP --> API
    CLAUDE <--> MCP

    API --> RETRIEVAL
    RETRIEVAL --> SEMANTIC
    RETRIEVAL --> KEYWORD
    RETRIEVAL --> RELATION

    SEMANTIC --> VECTOR
    KEYWORD --> FULLTEXT
    RELATION --> RELATIONS
    DOCSTORE --> RETRIEVAL
    CHUNKSTORE --> RETRIEVAL

    SEMANTIC --> HYBRID
    KEYWORD --> HYBRID
    RELATION --> HYBRID
    HYBRID --> RAG
    RAG --> PROMPT
    PROMPT --> LLM
    LLM --> RESPONSE
    RESPONSE --> API


## 2.Ingestion pipeline — sequence diagram
This describes what happens when a repository is initially indexed or when new commits arrive.

sequenceDiagram
    autonumber

    participant Scheduler as Scheduler / CLI
    participant Git as Git Repository
    participant Discovery as File Discovery
    participant Parser as Language Parsers
    participant Chunker as Chunking & Extraction
    participant Enrichment as LLM Enrichment
    participant Embedding as Embedding Provider
    participant DB as PostgreSQL

    Scheduler->>Git: Fetch latest commit
    Git-->>Scheduler: Commit SHA and changes

    Scheduler->>Discovery: Process changed files
    Discovery->>DB: Check existing file hashes
    DB-->>Discovery: Existing source metadata

    loop Each changed file
        Discovery->>Parser: Parse file
        Parser->>Parser: Identify symbols and structure
        Parser->>Chunker: Extract procedures / functions / sections

        Chunker->>Chunker: Create structured chunks
        Chunker->>Chunker: Extract line numbers and metadata
        Chunker->>DB: Save source, symbols and chunks

        opt Semantic enrichment enabled
            Chunker->>Enrichment: Submit selected chunks
            Enrichment->>Enrichment: Explain purpose and identify concepts
            Enrichment-->>Chunker: Enrichment and candidate relationships
            Chunker->>DB: Save reviewed/provisional enrichment
        end

        Chunker->>DB: Save structural relationships

        Chunker->>Embedding: Generate embeddings
        Embedding-->>Chunker: Vector embeddings
        Chunker->>DB: Store vectors and searchable content
    end

    Scheduler->>DB: Record ingestion run and commit SHA
    DB-->>Scheduler: Ingestion completed


## 3. Developer question — query and answer sequence
This is the runtime path when someone asks a question through the chat UI or CLI.


sequenceDiagram
    autonumber

    actor Developer
    participant UI as Chat UI / CLI
    participant API as FastAPI
    participant Assistant as Assistant Service
    participant Retrieval as Retrieval Service
    participant PG as PostgreSQL
    participant LLM as LLM Provider

    Developer->>UI: Ask a product / implementation question
    UI->>API: Submit question
    API->>Assistant: Process question

    Assistant->>Assistant: Identify intent and query terms

    Assistant->>Retrieval: Retrieve relevant knowledge

    par Semantic retrieval
        Retrieval->>PG: Vector similarity search
        PG-->>Retrieval: Relevant semantic chunks
    and Keyword retrieval
        Retrieval->>PG: Full-text / exact identifier search
        PG-->>Retrieval: Keyword matches
    and Relationship retrieval
        Retrieval->>PG: Find related symbols and dependencies
        PG-->>Retrieval: Related entities and source references
    end

    Retrieval->>Retrieval: Merge and rank results
    Retrieval->>PG: Fetch complete symbols and source metadata
    PG-->>Retrieval: Full content, commit and line references

    Retrieval-->>Assistant: Ranked evidence set

    Assistant->>Assistant: Build grounded context and citations
    Assistant->>LLM: Question + retrieved evidence
    LLM-->>Assistant: Generated answer

    Assistant->>Assistant: Validate / attach source references
    Assistant-->>API: Answer with citations
    API-->>UI: Return response
    UI-->>Developer: Display answer and source links


## 4. Claude Code integration — MCP sequence
This is an important part of the longer-term design. The goal is to make existing coding assistants product-aware rather than build a competing coding agent.


sequenceDiagram
    autonumber

    actor Developer
    participant Claude as Claude Code
    participant MCP as CLO Knowledge MCP Server
    participant API as Knowledge API
    participant Retrieval as Retrieval Service
    participant DB as PostgreSQL
    participant LLM as Knowledge LLM

    Developer->>Claude: Implement facility maintenance
    Claude->>MCP: search_product_knowledge("facility maintenance")
    MCP->>API: Search request
    API->>Retrieval: Retrieve domain knowledge
    Retrieval->>DB: Search domain documentation
    DB-->>Retrieval: Definitions and related concepts
    Retrieval-->>API: Evidence
    API-->>MCP: Search results
    MCP-->>Claude: Domain context and citations

    Claude->>MCP: search_legacy_code("facility maintenance")
    MCP->>API: Legacy code search
    API->>Retrieval: Search legacy repository
    Retrieval->>DB: Search symbols and chunks
    DB-->>Retrieval: ASP pages and SQL procedures
    Retrieval-->>API: Code evidence
    API-->>MCP: Results with file and line references
    MCP-->>Claude: Legacy implementation

    Claude->>MCP: search_modern_code("facility maintenance")
    MCP->>API: Modern code search
    API->>Retrieval: Search modern repository
    Retrieval->>DB: Search modern implementation
    DB-->>Retrieval: Services and contracts
    Retrieval-->>API: Code evidence
    API-->>MCP: Results
    MCP-->>Claude: Modern implementation

    Claude->>MCP: get_dependencies("FacilityService")
    MCP->>API: Dependency lookup
    API->>Retrieval: Expand relationships
    Retrieval->>DB: Query relationships
    DB-->>Retrieval: Callers, callees and tables
    Retrieval-->>API: Dependency graph
    API-->>MCP: Results
    MCP-->>Claude: Dependency information

    Claude->>LLM: Reason about implementation using retrieved context
    LLM-->>Claude: Design proposal
    Claude-->>Developer: Explanation, design and proposed code


## 5. Deployment architecture
The logical architecture above can run locally first and later be deployed using your existing Kubernetes infrastructure.


flowchart TB
    subgraph DEV["Developer Environment"]
        CLI["CLI"]
        CHAT["React Chat UI"]
        CLAUDE["Claude Code"]
    end

    subgraph INTERNAL["Internal Network / Kubernetes"]
        subgraph GATEWAY["Internal Access Layer"]
            GW["Internal Gateway"]
            AUTH["Authentication / Authorization<br/>Keycloak integration"]
        end

        subgraph APP["Knowledge Platform"]
            API["Knowledge API<br/>FastAPI"]
            MCP["MCP Server"]
            WORKER["Ingestion Worker<br/>Scheduled / CI-triggered"]
        end

        subgraph DBLAYER["Persistence"]
            PG[("PostgreSQL<br/>pgvector + Full-text")]
        end

        subgraph SOURCES["Source Access"]
            GIT["Legacy Git Repository"]
            GIT2["Modern Git Repository"]
            DOCS["Documentation Repository"]
        end

        subgraph OBS["Observability"]
            LOG["Logs"]
            METRICS["Metrics"]
            TRACE["Traces"]
        end
    end

    subgraph PROVIDERS["Approved AI Services"]
        EMB["Embedding Provider"]
        LLM["LLM Provider"]
    end

    CLI --> GW
    CHAT --> GW
    CLAUDE --> MCP
    MCP --> GW

    GW --> AUTH
    AUTH --> API
    GW --> MCP
    MCP --> API

    WORKER --> GIT
    WORKER --> GIT2
    WORKER --> DOCS
    WORKER --> EMB
    WORKER --> PG

    API --> PG
    API --> LLM
    API --> LOG
    API --> METRICS
    API --> TRACE
    WORKER --> LOG
    WORKER --> METRICS

    classDef external fill:#eef2ff,stroke:#6366f1,color:#1e1b4b
    classDef data fill:#ecfdf5,stroke:#059669,color:#064e3b
    class PG data
    class EMB,LLM external
