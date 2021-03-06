#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#include <omp.h>

#define RGB_COMPONENT_COLOR 255

typedef struct {
    unsigned char red, green, blue;
} PPMPixel;

typedef struct {
    int x, y;
    PPMPixel *data;
} PPMImage;

static PPMImage *readPPM() {
    char buff[16];
    PPMImage *img;
    FILE *fp;
    int c, rgb_comp_color;
    fp = stdin;
    
    if (!fgets(buff, sizeof(buff), fp)) {
        perror("stdin");
        exit(1);
    }
    
    if (buff[0] != 'P' || buff[1] != '6') {
        fprintf(stderr, "Invalid image format (must be 'P6')\n");
        exit(1);
    }
    
    img = (PPMImage *) malloc(sizeof(PPMImage));
    if (!img) {
        fprintf(stderr, "Unable to allocate memory\n");
        exit(1);
    }
    
    c = getc(fp);
    while (c == '#') {
        while (getc(fp) != '\n')
            ;
        c = getc(fp);
    }
    
    ungetc(c, fp);
    if (fscanf(fp, "%d %d", &img->x, &img->y) != 2) {
        fprintf(stderr, "Invalid image size (error loading)\n");
        exit(1);
    }
    
    if (fscanf(fp, "%d", &rgb_comp_color) != 1) {
        fprintf(stderr, "Invalid rgb component (error loading)\n");
        exit(1);
    }
    
    if (rgb_comp_color != RGB_COMPONENT_COLOR) {
        fprintf(stderr, "Image does not have 8-bits components\n");
        exit(1);
    }
    
    while (fgetc(fp) != '\n')
        ;
    img->data = (PPMPixel*) malloc(img->x * img->y * sizeof(PPMPixel));
    
    if (!img) {
        fprintf(stderr, "Unable to allocate memory\n");
        exit(1);
    }
    
    if (fread(img->data, 3 * img->x, img->y, fp) != img->y) {
        fprintf(stderr, "Error loading image.\n");
        exit(1);
    }
    
    return img;
}


void Histogram(PPMImage *image, float *h) {
    
    int i, j,  k, l, x, count;
    int rows, cols;    
    float n = image->y * image->x;
    int n_threads = omp_get_num_procs();//__builtin_omp_get_num_threads();

    cols = image->x;
    rows = image->y;

#pragma omp parallel for //schedule(dynamic) //schedule(dynamic,100)//schedule(static, chunk) //num_threads(n_threads)
    for (i = 0; i< (int)n; i++){
        //printf("Before = i %d = red %d blue %d green %d\n",i, image->data[i].red, image->data[i].blue,image->data[i].green);
        image->data[i].red = floor((image->data[i].red * 4) / 256);
        image->data[i].blue = floor((image->data[i].blue * 4) / 256);
        image->data[i].green = floor((image->data[i].green * 4) / 256);
        //printf("After = i %d = red %d blue %d green %d\n",i, image->data[i].red, image->data[i].blue,image->data[i].green);
    }
    
    count = 0;
    x = 0;
    int val;
    for (j = 0; j <= 3; j++) {
        for (k = 0; k <= 3; k++) {
            for (l = 0; l <= 3; l++) {
// Example from : https://stackoverflow.com/questions/12754485/openmp-custom-reduction-variable
                #pragma omp parallel //private(val) //num_threads(n_threads)//private(val)
                    {
                        count = 0;
                        int val = 0;  // val can be declared as local variable (for each thread) 
                        #pragma omp for //nowait       // now pragma for  (here you don't need to create threads, that's why no "omp parallel" )
                            // nowait specifies that the threads don't need to wait (for other threads to complete) after for loop, the threads can go ahead and execute the critical section 
                        for (i = 0; i < (int)n ; i++) {
                            if (image->data[i].red == j && image->data[i].green == k && image->data[i].blue == l) {
                                val++;
                            }
                        }
                        #pragma omp critical //atomic
                                count += val;
                    } 
                h[x] = count / n;
                count = 0;
                x++;
            }
        }
    }
}



int main(int argc, char *argv[]) {

     //int n_threads = omp_get_num_procs();
     //omp_set_num_threads(n_threads); 
    int i;
    
    PPMImage *image = readPPM();
    
    float *h = (float*)malloc(sizeof(float) * 64);
    
    for(i=0; i < 64; i++) h[i] = 0.0;
    
    Histogram(image, h);
    
    for (i = 0; i < 64; i++){
        printf("%0.3f ", h[i]);
    }
    printf("\n");
    free(h);
    
    return 0;
}
