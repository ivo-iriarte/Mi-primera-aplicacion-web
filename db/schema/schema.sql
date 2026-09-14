

CREATE TABLE institucion (
    id_institucion serial NOT NULL,
    nombre  varchar(150) NOT NULL,
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

CREATE TABLE deporte(
    id_deporte serial NOT NULL,
    nombre varchar(150) NOT NULL UNIQUE,
    activo boolean NOT NULL DEFAULT TRUE, 
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),

    CONSTRAINT PK_DEPORTE PRIMARY KEY(id_deporte)
);


CREATE TABLE cancha (
id_cancha serial NOT NULL,
nombre  varchar(150) NOT NULL,
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
        REFERENCES institucion(id_institucion),

    CONSTRAINT FK_CANCHA_DEPORTE 
        FOREIGN KEY (id_deporte) 
        REFERENCES deporte (id_deporte),

    CONSTRAINT chk_cancha_precio
        CHECK (precio >= 0),

    CONSTRAINT chk_cancha_duracion
        CHECK (duracion_turno_minutos > 0),

    CONSTRAINT chk_cancha_horario
        CHECK (hora_cierre > hora_apertura),

    CONSTRAINT uq_cancha_institucion_nombre
        UNIQUE (id_institucion, nombre)
);