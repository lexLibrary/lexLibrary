// Copyright (c) 2017 Townsourced Inc.

var gulp = require('gulp');
var path = require('path');

// var sass = require('gulp-sass');


var deployDir = './deploy';

// gulp.task('js', function (callback) {
//     // rollup
//     // buble
//     // copy vue dist to static
// });


// TODO: bulma
// gulp.task('css', function () {
//     return gulp.src('./src/sass/**/*.scss')
//         .pipe(sass({
//             outputStyle: 'compressed',
//             includePaths: 'node_modules'
//         }).on('error', sass.logError))
//         .pipe(gulp.dest(path.join(staticDir, 'css')))
//         .pipe(gulp.dest(path.join(cordovaDir, 'css')));
// });



// static files
gulp.task('html', function () {
    return gulp.src([
        './**/*.html',
        '!deploy/**/*',
        '!node_modules/**/*'
    ]).pipe(gulp.dest(deployDir));
});

gulp.task('images', function () {
    return gulp.src('./images/*')
        .pipe(gulp.dest(path.join(deployDir, 'images')))
});


// watch for changes
gulp.task('watch', function () {
    gulp.watch('./**/*.html', ['html']);
    gulp.watch('./images/**/*', ['images']);
    // gulp.watch(['./src/sass/**/*.scss'], ['css']);
    // gulp.watch(['./src/ts/**/*.ts', './src/ts/**/*.vue'], ['js']);
});


// start default task
gulp.task('default', ['html', 'images']);


//TODO: Production build task
