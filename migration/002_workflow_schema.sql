-- Workflow table (aggregate root)
CREATE TABLE IF NOT EXISTS workflow (
    id VARCHAR(26) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for listing workflows by creation date
CREATE INDEX IF NOT EXISTS idx_workflow_created_at ON workflow(created_at DESC);

-- NodeDefinition table (child entity of Workflow)
CREATE TABLE IF NOT EXISTS node_definition (
    id VARCHAR(26) PRIMARY KEY,
    workflow_id VARCHAR(26) NOT NULL,
    node_template_id VARCHAR(26) NOT NULL,
    name VARCHAR(255) NOT NULL,
    position_x DOUBLE PRECISION NOT NULL DEFAULT 0,
    position_y DOUBLE PRECISION NOT NULL DEFAULT 0,

    CONSTRAINT fk_node_definition_workflow
        FOREIGN KEY (workflow_id) REFERENCES workflow(id) ON DELETE CASCADE,
    CONSTRAINT fk_node_definition_node_template
        FOREIGN KEY (node_template_id) REFERENCES node_template(id)
);

-- Index for finding node definitions by workflow
CREATE INDEX IF NOT EXISTS idx_node_definition_workflow_id ON node_definition(workflow_id);

-- Edge table (child entity of Workflow, connects NodeDefinitions)
CREATE TABLE IF NOT EXISTS edge (
    id VARCHAR(26) PRIMARY KEY,
    workflow_id VARCHAR(26) NOT NULL,
    from_node_id VARCHAR(26) NOT NULL,
    to_node_id VARCHAR(26) NOT NULL,

    CONSTRAINT fk_edge_workflow
        FOREIGN KEY (workflow_id) REFERENCES workflow(id) ON DELETE CASCADE,
    CONSTRAINT fk_edge_from_node
        FOREIGN KEY (from_node_id) REFERENCES node_definition(id) ON DELETE CASCADE,
    CONSTRAINT fk_edge_to_node
        FOREIGN KEY (to_node_id) REFERENCES node_definition(id) ON DELETE CASCADE,

    -- Prevent duplicate edges between same nodes
    CONSTRAINT uq_edge_from_to UNIQUE (workflow_id, from_node_id, to_node_id)
);

-- Index for finding edges by workflow
CREATE INDEX IF NOT EXISTS idx_edge_workflow_id ON edge(workflow_id);
