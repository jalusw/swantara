import { test } from "@playwright/test";

test("webgl2 screenshot sanity check", async ({ page }) => {
  await page.setContent(
    `<canvas id="c" width="256" height="256"></canvas>
     <script>
       const gl = document.getElementById('c').getContext('webgl2');
       const vs = gl.createShader(gl.VERTEX_SHADER); gl.shaderSource(vs, "#version 300 es\nin vec2 p; void main(){gl_Position=vec4(p,0.,1.);}"); gl.compileShader(vs);
       const fs = gl.createShader(gl.FRAGMENT_SHADER); gl.shaderSource(fs, "#version 300 es\nprecision mediump float; out vec4 o; void main(){o=vec4(0.,1.,0.,1.);}"); gl.compileShader(fs);
       const prog = gl.createProgram(); gl.attachShader(prog, vs); gl.attachShader(prog, fs); gl.linkProgram(prog); gl.useProgram(prog);
       const buf = gl.createBuffer(); gl.bindBuffer(gl.ARRAY_BUFFER, buf); gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1,-1, 1,-1, -1,1, 1,1]), gl.STATIC_DRAW);
       const loc = gl.getAttribLocation(prog, "a"); gl.enableVertexAttribArray(loc); gl.vertexAttribPointer(loc, 2, gl.FLOAT, false, 0, 0);
       gl.clearColor(0,0,0,1); gl.clear(gl.COLOR_BUFFER_BIT); gl.drawArrays(gl.TRIANGLE_STRIP, 0, 4);
     </script>`,
  );
  await page.waitForTimeout(1000);
  await page.screenshot({ path: "/tmp/opencode/webgl2-sanity.png" });
});

test("webgl1 canvas after DOM ready in design page", async ({ page }) => {
  await page.goto("http://localhost:3000/preview/design-system", {
    waitUntil: "domcontentloaded",
    timeout: 90000,
  });
  await page.waitForTimeout(12000);
  await page.screenshot({ path: "/tmp/opencode/design-top.png" });
});
