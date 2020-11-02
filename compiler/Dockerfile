FROM emscripten/emsdk:2.0.4

WORKDIR /cfimager_compiler
RUN chown 500 /cfimager_compiler
USER 500

RUN mkdir src
RUN mkdir objs

COPY --chown=500 ./core/bin/*.o objs/
COPY --chown=500 core/flib/flib.h src/
COPY --chown=500 ./compiler .

EXPOSE 8080/tcp

CMD ["./compiler"]