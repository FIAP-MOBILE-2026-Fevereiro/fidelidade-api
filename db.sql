-- Tabela de Usuários
CREATE TABLE usuarios (
    id VARCHAR(12) PRIMARY KEY, -- Formato: usr_XXXXXXXX
    nome VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    senha_hash VARCHAR(255) NOT NULL,
    imagem_perfil VARCHAR(500),
    ativo BOOLEAN DEFAULT true,
    data_criacao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    data_atualizacao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_ativo (ativo)
);

-- Tabela de Programas de Fidelidade
CREATE TABLE programas (
    id VARCHAR(13) PRIMARY KEY, -- Formato: prog_XXXXXXXX
    estabelecimento_id VARCHAR(12) NOT NULL,
    nome_estabelecimento VARCHAR(200) NOT NULL,
    lat DECIMAL(9,6) NOT NULL,
    lng DECIMAL(9,6) NOT NULL,
    selos_necessarios INTEGER NOT NULL CHECK (selos_necessarios > 0),
    recompensa_nome VARCHAR(200) NOT NULL,
    recompensa_imagem VARCHAR(500),
    recompensa_descricao TEXT,
    descricao TEXT,
    regras TEXT,
    ativo BOOLEAN DEFAULT true,
    data_inicio TIMESTAMP NOT NULL,
    data_fim TIMESTAMP,
    data_criacao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_localizacao (lat, lng),
    INDEX idx_ativo (ativo),
    SPATIAL INDEX idx_geo (lat, lng)
);

-- Tabela de Selos
CREATE TABLE selos (
    id VARCHAR(13) PRIMARY KEY, -- Formato: selo_XXXXXXXX
    usuario_id VARCHAR(12) NOT NULL,
    programa_id VARCHAR(13) NOT NULL,
    data_aquisicao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    chave_validacao VARCHAR(255) UNIQUE NOT NULL,
    sequencia INTEGER NOT NULL,
    validado BOOLEAN DEFAULT true,
    qr_code_hash VARCHAR(255) UNIQUE NOT NULL,
    FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE,
    FOREIGN KEY (programa_id) REFERENCES programas(id) ON DELETE CASCADE,
    INDEX idx_usuario_programa (usuario_id, programa_id),
    INDEX idx_programa (programa_id),
    INDEX idx_data_aquisicao (data_aquisicao),
    UNIQUE KEY uk_usuario_programa_data (usuario_id, programa_id, DATE(data_aquisicao))
);

-- Tabela de QR Codes gerados pelos estabelecimentos
CREATE TABLE qr_codes (
    id VARCHAR(12) PRIMARY KEY,
    programa_id VARCHAR(13) NOT NULL,
    estabelecimento_id VARCHAR(12) NOT NULL,
    codigo_unico VARCHAR(255) UNIQUE NOT NULL,
    data_geracao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    data_expiracao TIMESTAMP NOT NULL,
    utilizado BOOLEAN DEFAULT false,
    usuario_id VARCHAR(12),
    data_utilizacao TIMESTAMP,
    FOREIGN KEY (programa_id) REFERENCES programas(id) ON DELETE CASCADE,
    FOREIGN KEY (usuario_id) REFERENCES usuarios(id),
    INDEX idx_programa_codigo (programa_id, codigo_unico),
    INDEX idx_expiracao (data_expiracao)
);

-- Tabela de Recompensas Resgatadas
CREATE TABLE recompensas_resgatadas (
    id VARCHAR(12) PRIMARY KEY,
    usuario_id VARCHAR(12) NOT NULL,
    programa_id VARCHAR(13) NOT NULL,
    codigo_resgate VARCHAR(50) UNIQUE NOT NULL,
    data_resgate TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    data_expiracao_resgate TIMESTAMP,
    resgatada_no_estabelecimento BOOLEAN DEFAULT false,
    data_confirmacao_resgate TIMESTAMP,
    FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE,
    FOREIGN KEY (programa_id) REFERENCES programas(id) ON DELETE CASCADE,
    INDEX idx_usuario_resgate (usuario_id, programa_id)
);