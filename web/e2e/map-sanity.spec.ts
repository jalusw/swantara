import { test } from "@playwright/test";

test("webgl screenshot sanity check", async ({ page }) => {
  await page.setContent(
    `<canvas id="c" width="256" height="256"></canvas>
     <script>
       const gl = document.getElementById('c').getContext('webgl2');
       // red full-screen quad
       const vs = gl.createShader(gl.VERTEX_SHADER); gl.shaderSource(vs, 'attribute vec2 p; void main(){gl_Position=vec4(p,0.,1.);}'); gl.compileShader(vs);
       const fs = gl.createShader(gl.FRAGMENT_SHADER); gl.shaderSource(fs, 'precision mediump float; void main(){gl_FragColor=vec4(1.,0.,0.,1.);}'); gl.compileShader(fs);
       const prog = gl.createProgram(); gl.attachShader(prog, vs); gl.attachShader(prog, fs); gl.linkProgram(prog); gl.useProgram(prog);
       const buf = gl.createBuffer(); gl.bindBuffer(gl.ARRAY_BUFFER, buf); gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1,-1, 1,-1, -1,1, 1,1]), gl.STATIC_DRAW);
       const loc = gl.getAttribLocation(prog, "p"); gl.enableVertexAttribArray(loc); gl.vertexAttribPointer(loc, 2, gl.FLOAT, false, 0, 0);
       gl.clearColor(0,0,0,1); gl.clear(gl.COLOR_BUFFER_BIT); gl.drawArrays(gl.TRIANGLE_STRIP, 0, 4);
     </script>`,
  );
  await page.waitForTimeout(1000);
  const box = await page.locator("#c").boundingBox();
  console.log("canvas box:", JSON.stringify(box));
  await page.screenshot({ path: "/tmp/opencode/webgl-sanity.png" });
});
