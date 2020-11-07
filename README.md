## THIS IS A PROOF OF CONCEPT, EXPECT MANY FEATURES TO BE MISSING, AND BUGS TO OCCUR

### What it does
Let's assume we have a function `f` and `f(1 + i) = 0.5 + 0.5i`.

We could place an image onto the complex plane, and move the pixel which is over the `1 + i` point to the `0.5 + 0.5i` point.

This way, we can get an intuition of what this function does: we can see which
 zones go where, or wether it moves the image, rotates it, does both, zooms out...

See below for imaged examples.

### Function example
This is the most basic function example: do nothing.
```c
#include <complex.h> // get library to do complex number computation
// see https://www.gnu.org/software/libc/manual/html_node/Complex-Numbers.html for documentation

// f is the function we will draw
// it takes a complex number c as input : the origin point
// it must return a complex number: the destination point
complex double f(complex double c)
{ // edit below
    return c; // return the same position, unchanged
}
```

### Using a function
This project provides a web UI through which you can create a use functions.
A live demo can be found [here](https://cfimager.neirpyc.ovh). 
You have to register, so you can create your own functions, you can then create a function, 
compile it, and use it.

### Parameters
The render page provides multiple options:
- `inputRange`: a complex range which represents where the image is placed on the complex plane before transformation.
  It is the equivalent of `xMin` and `xMax` for a real function.
- `outputRange`: a complex range which represents the part of the complex plane after transformation which is displayed.
  It is the equivalent of `yMin` and `yMax` for a real function.
- `outputDimension`: width and height of the output image.
- `samplingFactor`: a float which represents the number of points moved per pixel in the output image.
- `handleCollisions`: whether or color blending should occur when multiple point land on the same pixel.
  Render is slightly slower when enabled, but often yields a much better result.

A complex range is a string of the following format: `{%f%fi;%f%fi}` with `%f` being a `+` or a `-`
followed by at least one digit, which may be followed by a `.` and at least one digit.
  
### License
This poject is published under a GPL v3 license.

See [LICENSE](./LICENSE).

### Example of results 
- Source image: ![source image](https://cfimager.neirpyc.ovh/githubImages/input.jpg)
- `f(a + ib) = a + 0.5 + 0.5i`: ![shifted image](https://cfimager.neirpyc.ovh/githubImages/cst.jpg)
- `f(a + ib) = (a + ib) * 1.41`: ![scaled image](https://cfimager.neirpyc.ovh/githubImages/scale.jpg)
- `f(a + ib) = (a + ib) * i * 0.75`: ![roated image](https://cfimager.neirpyc.ovh/githubImages/rot.jpg)
- Riemann's zeta function: ![weird image](https://cfimager.neirpyc.ovh/githubImages/zeta.jpg)

### Reporting a bug
If you think you have found a bug, you can help us fixing it faster by opening a github issue
and providing the following informations:

```md
Browser: <browser name> <browser version>
Email: <email used to register>
Page: <url of the page(s) on which the bug occured>
Function (if relevant): <source code of the function you used>

## What happened:

## What should have happened:
```

### Feature request
If you would like a feature to be included in this project, feel free to open an issue, 
and describe the missing feature.

### Contributing

There are many ways you can contribute to this project, even without knowing how to code.
 - You know about web design, css and stuff like that? Our UI is terrible, you can open pull 
 requests to improve it or get in touch, so we can talk :)
 - You know about color blending and alpha blending? I'm terrible at it, and I must have done 
 things wrong, and I'd love to learn more about how I should do this!
 - You know about IT security? This project is about compiling untrusted code and 
 user authentication! It's likely I've done things wrong or left an obvious XSS.
 - You know about math? I'm sure you can make rendering faster, or maybe you have idea of pretty 
 functions we can showcase in this README?
 - You speak english better than me? I'm french, I must have made many mistakes, open
 a pull request and improve this README.
 - You're a C/GO/JS developer? Find //todos and fix them, implement your own features
 , or get in touch!
 - You're crazy? Go ahead and comment my code. Just kidding, I'll do it, but I first 
 wanted to get a working proof of concept before doing things the proper way!
 - You've got an extra server? Mine is super slow, compiling functions can take up to a minute, 
 with a faster one, this could go under 20 seconds!

And if you think you can contribute to this project, open a PR, an issue, or get in touch!