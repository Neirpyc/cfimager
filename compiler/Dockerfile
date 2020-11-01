FROM emscripten/emsdk:2.0.4 as builder
WORKDIR /cfimager_compiler

RUN mkdir objs
COPY core core
RUN /bin/sh -c "cd core && make fclean && make prepare && cd .."
RUN cp core/*.o objs/
RUN cp core/core/conversion/*.o objs/
RUN cp core/core/image/*.o objs/
RUN cp core/core/process/*.o objs/

FROM emscripten/emsdk:2.0.4

WORKDIR /cfimager_compiler
RUN chown 500 /cfimager_compiler
USER 500

RUN mkdir src
RUN mkdir objs

COPY --from=builder --chown=500 /cfimager_compiler/objs objs/
COPY --chown=500 core/flib/flib.h src/
COPY --chown=500 ./compiler .

EXPOSE 8080/tcp

CMD ["./compiler"]