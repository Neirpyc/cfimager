NAME = cfimager.js

SRC = main.c flib/flib.c set_render_settings.c core/conversion/conversion.c  \
core/image/image.c  core/process/process.c core/process/async.c


OBJS = $(SRC:.c=.o)
OBJS_NO_FLIB = $(filter-out flib/flib.o, $(OBJS))

RM = rm -f
CC = emcc
CFLAGS = -s EXTRA_EXPORTED_RUNTIME_METHODS='["cwrap", "ccall"]' -s USE_PTHREADS=1 \
-s EXPORTED_FUNCTIONS='["_main", "_draw", "_stop_render", "_set_render_settings"]' -O3 -s EXPORT_ALL=0 \
-s ALLOW_MEMORY_GROWTH=1 -s ASSERTIONS=1


$(NAME): all

all: $(OBJS)
	$(CC) $(CFLAGS) $(OBJS) -o $(NAME)

clean:
	$(RM) $(OBJS)

fclean: clean
	$(RM) $(NAME) $(NAME:.js=.wasm) $(NAME:.js=.worker.js)

re: clean all

cleanflib:
	$(RM) flib/flib.o

prepare: cleanflib $(OBJS_NO_FLIB)

%.o: %.c
	$(CC) $(CFLAGS) -c $< -o $@