
-- Tablas independientes 

CREATE TABLE institucion (
    id_institucion serial NOT NULL,
    nombre varchar(150) NOT NULL,
    direccion varchar NOT NULL,
    telefono varchar(30) NOT NULL,
    email varchar NOT NULL,
    descripcion text NOT NULL,
    activo boolean NOT NULL DEFAULT TRUE,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    ciudad varchar(150) NOT NULL,
    provincia varchar(150) NOT NULL,

    CONSTRAINT PK_INSTITUCION PRIMARY KEY (id_institucion)
);

CREATE TABLE deporte (
    id_deporte serial NOT NULL,
    nombre varchar(100) NOT NULL UNIQUE,
    activo boolean NOT NULL DEFAULT TRUE,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),

    CONSTRAINT PK_DEPORTE PRIMARY KEY (id_deporte)
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

    CONSTRAINT PK_USUARIO PRIMARY KEY (id_usuario)
);

CREATE TABLE estados_reservas (
    id_estados_reserva serial NOT NULL,
    nombre varchar(150) NOT NULL UNIQUE,

    CONSTRAINT PK_ESTADOS_RESERVAS PRIMARY KEY (id_estados_reserva)
);

CREATE TABLE servicios (
    id_servicio serial NOT NULL,
    servicio varchar(150) NOT NULL UNIQUE,
    descripcion text,
    activo boolean NOT NULL DEFAULT TRUE,

    CONSTRAINT PK_SERVICIOS PRIMARY KEY (id_servicio)
);

-- tablas dependientes

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

    CONSTRAINT chk_cancha_precio
        CHECK (precio >= 0),

    CONSTRAINT chk_cancha_duracion
        CHECK (duracion_turno_minutos > 0),

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


--prueba de indices vistos en la cursada de BD2 ya que estamos no vienen de mas 
CREATE INDEX idx_reservas_cancha_fechas ON reservas (id_cancha, fecha_inicio, fecha_fin);
CREATE INDEX idx_reservas_usuario ON reservas (id_usuario);
CREATE INDEX idx_cancha_institucion ON cancha (id_institucion);