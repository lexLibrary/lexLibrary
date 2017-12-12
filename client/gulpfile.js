// Copyright (c) 2017 Townsourced Inc.

var gulp = require('gulp');
// var sass = require('gulp-sass');

gulp.task('js', function (callback) {
    // rollup
    // buble
    // copy veu dist to static
});


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
// gulp.task('html', function () {
//     return gulp.src('./src/*.html')
//         .pipe(gulp.dest(staticDir))
//         .pipe(gulp.dest(cordovaDir));
// });

// gulp.task('images', function () {
//     return gulp.src('./src/images/*')
//         .pipe(gulp.dest(path.join(staticDir, 'images')))
//         .pipe(gulp.dest(path.join(cordovaDir, 'images')));
// });


// watch for changes
gulp.task('watch', function () {
    // gulp.watch('./src/*.html', ['html']);
    // gulp.watch('./src/images/*', ['images']);
    // gulp.watch(['./src/sass/**/*.scss'], ['css']);
    // gulp.watch(['./src/ts/**/*.ts', './src/ts/**/*.vue'], ['js']);
});


// start default task
gulp.task('default', ['images', 'css', 'js', 'watch']);


//TODO: Production build task
