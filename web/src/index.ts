class Grid {
    private canvas: HTMLCanvasElement;
    private context: CanvasRenderingContext2D;
    private values: Boolean[][];

    constructor() {
        const numberOfSquare = 20
        let canvas = document.getElementById('game') as
            HTMLCanvasElement;
        let context = canvas.getContext("2d")!;
        canvas.height = 800
        canvas.width = 800
        let squareSize = canvas.height / numberOfSquare
        context.lineCap = 'round';
        context.lineJoin = 'round';
        context.strokeStyle = 'black';
        context.lineWidth = 1;
        canvas.style.border = '1px solid #000'

        this.canvas = canvas;
        this.context = context;
        for (var x = 0; x < canvas.height; x += squareSize) {
            context.moveTo(x, 0)
            context.lineTo(x, canvas.height)
            context.moveTo(0, x)
            context.lineTo(canvas.height, x)
            console.log(x)
        }
        context.strokeStyle = "black";
        context.stroke();
    }
}

new Grid();
