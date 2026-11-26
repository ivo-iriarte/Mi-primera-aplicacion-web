
-- Tablas independientes 

CREATE TABLE institucion (
    id_institucion serial NOT NULL,
    nombre varchar(150) NOT NULL,
    direccion varchar(255) NOT NULL,
    telefono varchar(30) NOT NULL,
    email varchar(255) NOT NULL UNIQUE,
    descripcion text NOT NULL DEFAULT '',
    activo boolean NOT NULL DEFAULT TRUE,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    ciudad varchar(150) NOT NULL,
    provincia varchar(150) NOT NULL,

    CONSTRAINT PK_INSTITUCION PRIMARY KEY (id_institucion),
    CONSTRAINT chk_institucion_nombre CHECK (length(trim(nombre)) > 0),
    CONSTRAINT chk_institucion_email CHECK (length(trim(email)) > 0)
);

CREATE TABLE deporte (
    id_deporte serial NOT NULL,
    nombre varchar(100) NOT NULL UNIQUE,
    activo boolean NOT NULL DEFAULT TRUE,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),

    CONSTRAINT PK_DEPORTE PRIMARY KEY (id_deporte),
    CONSTRAINT chk_deporte_nombre CHECK (length(trim(nombre)) > 0)
);

CREATE TABLE usuario (
    id_usuario serial NOT NULL,
    nombre varchar(100) NOT NULL,
    apellido varchar(100) NOT NULL,
    email varchar(255) NOT NULL UNIQUE,
    password_hash varchar(255) NOT NULL,
    telefono varchar(30),
    activo boolean NOT NULL DEFAULT TRUE,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),

    CONSTRAINT PK_USUARIO PRIMARY KEY (id_usuario),
    CONSTRAINT chk_usuario_nombre CHECK (length(trim(nombre)) > 0),
    CONSTRAINT chk_usuario_apellido CHECK (length(trim(apellido)) > 0),
    CONSTRAINT chk_usuario_email CHECK (length(trim(email)) > 0)
);

CREATE TABLE estados_reservas (
    id_estados_reserva serial NOT NULL,
    nombre varchar(150) NOT NULL UNIQUE,

    CONSTRAINT PK_ESTADOS_RESERVAS PRIMARY KEY (id_estados_reserva),
    CONSTRAINT chk_estado_reserva_nombre CHECK (length(trim(nombre)) > 0)
);

CREATE TABLE servicios (
    id_servicio serial NOT NULL,
    servicio varchar(150) NOT NULL UNIQUE,
    descripcion text,
    activo boolean NOT NULL DEFAULT TRUE,

    CONSTRAINT PK_SERVICIOS PRIMARY KEY (id_servicio),
    CONSTRAINT chk_servicio_nombre CHECK (length(trim(servicio)) > 0)
);

-- Tablas dependientes

CREATE TABLE cancha (
    id_cancha serial NOT NULL,
    nombre varchar(100) NOT NULL,
    precio decimal(10,2) NOT NULL,
    activa boolean NOT NULL DEFAULT TRUE,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    duracion_turno_minutos smallint NOT NULL,
    hora_apertura time NOT NULL,
    hora_cierre time NOT NULL,
    id_institucion integer NOT NULL,
    id_deporte integer NOT NULL,

    CONSTRAINT PK_CANCHA 
        PRIMARY KEY (id_cancha),

    CONSTRAINT FK_CANCHA_INSTITUCION 
        FOREIGN KEY (id_institucion) 
        REFERENCES institucion(id_institucion)
        ON DELETE RESTRICT,

    CONSTRAINT FK_CANCHA_DEPORTE 
        FOREIGN KEY (id_deporte) 
        REFERENCES deporte (id_deporte)
        ON DELETE RESTRICT,

    CONSTRAINT chk_cancha_nombre
        CHECK (length(trim(nombre)) > 0),

    CONSTRAINT chk_cancha_precio
        CHECK (precio >= 0),

    CONSTRAINT chk_cancha_duracion
        CHECK (duracion_turno_minutos > 0 AND duracion_turno_minutos <= 480),

    CONSTRAINT chk_cancha_horario
        CHECK (hora_cierre > hora_apertura),

    CONSTRAINT uq_cancha_institucion_nombre
        UNIQUE (id_institucion, nombre)
);

CREATE TABLE cancha_servicio (
    id_servicio integer NOT NULL,
    id_cancha integer NOT NULL,

    CONSTRAINT PK_CANCHA_SERVICIO 
        PRIMARY KEY (id_servicio, id_cancha),

    CONSTRAINT FK_CANCHA_SERVICIO_SERVICIO 
        FOREIGN KEY (id_servicio) 
        REFERENCES servicios(id_servicio)
        ON DELETE CASCADE,

    CONSTRAINT FK_CANCHA_SERVICIO_CANCHA 
        FOREIGN KEY (id_cancha) 
        REFERENCES cancha(id_cancha)
        ON DELETE CASCADE
);

CREATE TABLE reservas (
    id_reserva serial NOT NULL,
    id_usuario integer NOT NULL,
    id_cancha integer NOT NULL,
    fecha_inicio timestamptz NOT NULL,
    fecha_fin timestamptz NOT NULL,
    id_estado integer NOT NULL,
    precio_total decimal(10,2) NOT NULL,
    observaciones text,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),

    CONSTRAINT PK_RESERVAS 
        PRIMARY KEY (id_reserva),

    CONSTRAINT FK_RESERVAS_USUARIO 
        FOREIGN KEY (id_usuario) 
        REFERENCES usuario(id_usuario)
        ON DELETE RESTRICT,

    CONSTRAINT FK_RESERVAS_CANCHA 
        FOREIGN KEY (id_cancha) 
        REFERENCES cancha(id_cancha)
        ON DELETE RESTRICT,

    CONSTRAINT FK_RESERVAS_ESTADO 
        FOREIGN KEY (id_estado) 
        REFERENCES estados_reservas(id_estados_reserva)
        ON DELETE RESTRICT,

    CONSTRAINT chk_reserva_fechas 
        CHECK (fecha_fin > fecha_inicio),

    CONSTRAINT chk_reserva_precio 
        CHECK (precio_total >= 0)
);

-- Índices para acelerar las consultas de los flujos principales de la aplicación
CREATE INDEX idx_cancha_institucion ON cancha (id_institucion);
CREATE INDEX idx_cancha_deporte ON cancha (id_deporte);
CREATE INDEX idx_cancha_institucion_deporte ON cancha (id_institucion, id_deporte);
CREATE INDEX idx_cancha_servicio_cancha ON cancha_servicio (id_cancha);
CREATE INDEX idx_reservas_cancha_fechas ON reservas (id_cancha, fecha_inicio, fecha_fin);
CREATE INDEX idx_reservas_usuario ON reservas (id_usuario);
CREATE INDEX idx_reservas_estado ON reservas (id_estado);