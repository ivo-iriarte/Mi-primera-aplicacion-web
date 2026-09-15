# **Trabajo práctico**

Este proyecto fue desarrollado en Go utilizando PostgreSQL como base de datos. Para facilitar tanto la compilación como la ejecución de los tests, las tareas principales se encuentran automatizadas mediante un `Makefile`.

## **Cómo ejecutar el proyecto**

Primero se debe clonar el repositorio y entrar en la carpeta del proyecto:

git clone https://github.com/ivo-iriarte/Mi-primera-aplicacion-web.git

cd Mi-primera-aplicacion-web.git

Una vez dentro, se pueden compilar los archivos y ejecutar los tests con el siguiente comando:

make build && make test

Para poder realizar este proceso es necesario tener instalados Go, Docker y Docker Compose

## **Desarrollo**

Para el desarrollo se decidió trabajar principalmente con la tabla `Cancha`. Como una cancha pertenece a una institución y está asociada a un deporte, también fue necesario implementar las tablas `Institución` y `Deporte`.

Estas tablas se encuentran relacionadas mediante claves foráneas. De esta manera, cada registro de `Cancha` puede referenciar correctamente a la institución a la que pertenece y al deporte para el cual se utiliza, evitando guardar información sin una relación válida dentro de la base de datos. (directamente por como estan definidos los SQL no se podria crear los registros)

La definición de las tablas se realizó en el archivo de esquema SQL. Por otro lado, las operaciones necesarias para crear, consultar, actualizar y eliminar registros se escribieron en el archivo de queries.

A partir de los archivos de esquema y queries se utilizó `sqlc` para generar automáticamente los archivos `.go` encargados de interactuar con PostgreSQL. Esto permitió mantener las consultas SQL separadas del resto del código y trabajar con funciones y tipos generados para Go.

## **Base de datos para los tests**

Para ejecutar los tests se eligió levantar una instancia de PostgreSQL en un contenedor Docker separado del contenedor utilizado para la persistencia normal del proyecto.

Esta separación permite que las pruebas trabajen sobre una base de datos independiente, sin modificar ni eliminar los datos guardados durante el uso habitual de la aplicación. La base utilizada por los tests puede crearse desde cero y eliminarse al finalizar cada ejecución.

En lugar de utilizar un script independiente para automatizar este proceso, se decidió usar un `Makefile`. La elección se hizo con el objetivo de aprender cómo funciona esta herramienta y centralizar en un mismo lugar los comandos necesarios para compilar el proyecto, preparar el entorno de prueba y ejecutar los tests.

## **Reflexion**

A lo largo del proyecto nos pusimos a discutir bastante sobre si convenía casarse con un ORM o ir por SQL directo. Nos dimos cuenta de que, si bien las herramientas de abstracción te resuelven rapidísimo las cosas simples, en cuanto querés hacer consultas un poco más rebuscadas te terminan complicando la vida o escondiendo lo que pasa de fondo. Llegamos a la conclusión de que lo ideal es un equilibrio: usar abstracciones para no reinventar la rueda en lo básico, pero apoyarse en SQL puro —con herramientas como sqlc— cuando necesitás control total y claridad. Además, teniendo en cuenta que hoy en día con la asistencia de IA es mucho más rápido escribir y revisar consultas nativas, encararlo de esta manera nos vino perfecto para entender de verdad cómo dialogan Go y PostgreSQL, que al final era el gran objetivo que teníamos para aprender con este trabajo.
&nbsp;
