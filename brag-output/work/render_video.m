#import <Foundation/Foundation.h>
#import <CoreGraphics/CoreGraphics.h>
#import <CoreText/CoreText.h>
#import <ImageIO/ImageIO.h>
#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define WIDTH 1920
#define HEIGHT 1080
#define FPS 30
#define DURATION 24.0
#define TOTAL_FRAMES 720
#define POSTER_TIME 7.40

typedef struct {
    CGFloat r, g, b, a;
} RGBA;

static inline RGBA hexColor(uint32_t hex, CGFloat a) {
    return (RGBA){
        ((hex >> 16) & 0xFF) / 255.0,
        ((hex >> 8) & 0xFF) / 255.0,
        (hex & 0xFF) / 255.0,
        a
    };
}

static inline RGBA mixColor(RGBA c1, RGBA c2, CGFloat t) {
    if (t < 0) t = 0;
    if (t > 1) t = 1;
    return (RGBA){
        c1.r + (c2.r - c1.r) * t,
        c1.g + (c2.g - c1.g) * t,
        c1.b + (c2.b - c1.b) * t,
        c1.a + (c2.a - c1.a) * t
    };
}

static inline CGFloat clampf(CGFloat v, CGFloat lo, CGFloat hi) {
    return v < lo ? lo : (v > hi ? hi : v);
}

static inline CGFloat smoothstep(CGFloat edge0, CGFloat edge1, CGFloat x) {
    CGFloat t = clampf((x - edge0) / (edge1 - edge0), 0.0, 1.0);
    return t * t * (3.0 - 2.0 * t);
}

static inline CGFloat easeOutCubic(CGFloat x) {
    x = clampf(x, 0.0, 1.0);
    CGFloat inv = 1.0 - x;
    return 1.0 - inv * inv * inv;
}

static inline CGFloat easeOutBack(CGFloat x) {
    x = clampf(x, 0.0, 1.0);
    CGFloat c1 = 1.70158;
    CGFloat c3 = c1 + 1.0;
    return 1.0 + c3 * pow(x - 1.0, 3) + c1 * pow(x - 1.0, 2);
}

static inline CGFloat easeInCubic(CGFloat x) {
    x = clampf(x, 0.0, 1.0);
    return x * x * x;
}

static inline void sceneTransition(CGFloat t, CGFloat tStart, CGFloat tEnd, CGFloat *outAlpha, CGFloat *outOffsetY, CGFloat *outScale) {
    CGFloat enterDur = 0.32;
    CGFloat exitDur = 0.22;
    if (t < tStart || t >= tEnd) {
        *outAlpha = 0.0;
        *outOffsetY = 24.0;
        *outScale = 0.97;
        return;
    }
    if (t < tStart + enterDur) {
        CGFloat u = easeOutCubic((t - tStart) / enterDur);
        *outAlpha = u;
        *outOffsetY = (1.0 - u) * 26.0;
        *outScale = 0.97 + 0.03 * u;
    } else if (t > tEnd - exitDur) {
        CGFloat u = easeInCubic((t - (tEnd - exitDur)) / exitDur);
        *outAlpha = 1.0 - u;
        *outOffsetY = -u * 20.0;
        *outScale = 1.0 - 0.02 * u;
    } else {
        *outAlpha = 1.0;
        *outOffsetY = 0.0;
        *outScale = 1.0;
    }
}

static void setFill(CGContextRef ctx, RGBA c) {
    CGContextSetRGBFillColor(ctx, c.r, c.g, c.b, c.a);
}

static void setStroke(CGContextRef ctx, RGBA c, CGFloat w) {
    CGContextSetRGBStrokeColor(ctx, c.r, c.g, c.b, c.a);
    CGContextSetLineWidth(ctx, w);
}

static void addRoundedRectPath(CGContextRef ctx, CGRect r, CGFloat radius) {
    CGFloat maxR = fmin(r.size.width, r.size.height) * 0.5;
    if (radius > maxR) radius = maxR;
    CGPathRef path = CGPathCreateWithRoundedRect(r, radius, radius, NULL);
    CGContextAddPath(ctx, path);
    CGPathRelease(path);
}

static void fillRoundRect(CGContextRef ctx, CGRect r, CGFloat radius, RGBA fill) {
    CGContextSaveGState(ctx);
    setFill(ctx, fill);
    addRoundedRectPath(ctx, r, radius);
    CGContextFillPath(ctx);
    CGContextRestoreGState(ctx);
}

static void strokeRoundRect(CGContextRef ctx, CGRect r, CGFloat radius, RGBA stroke, CGFloat lineWidth) {
    CGContextSaveGState(ctx);
    setStroke(ctx, stroke, lineWidth);
    addRoundedRectPath(ctx, r, radius);
    CGContextStrokePath(ctx);
    CGContextRestoreGState(ctx);
}

static void fillAndStrokeRoundRect(CGContextRef ctx, CGRect r, CGFloat radius, RGBA fill, RGBA stroke, CGFloat lineWidth) {
    CGContextSaveGState(ctx);
    addRoundedRectPath(ctx, r, radius);
    setFill(ctx, fill);
    CGContextFillPath(ctx);
    addRoundedRectPath(ctx, r, radius);
    setStroke(ctx, stroke, lineWidth);
    CGContextStrokePath(ctx);
    CGContextRestoreGState(ctx);
}

static void drawRadialGlow(CGContextRef ctx, CGFloat cx, CGFloat cy, CGFloat radius, RGBA color) {
    if (radius <= 0 || color.a <= 0.001) return;
    CGContextSaveGState(ctx);
    CGColorSpaceRef cs = CGColorSpaceCreateDeviceRGB();
    CGFloat comps[8] = {
        color.r, color.g, color.b, color.a,
        color.r, color.g, color.b, 0.0
    };
    CGFloat locs[2] = { 0.0, 1.0 };
    CGGradientRef grad = CGGradientCreateWithColorComponents(cs, comps, locs, 2);
    CGContextDrawRadialGradient(ctx, grad, CGPointMake(cx, cy), 0, CGPointMake(cx, cy), radius, 0);
    CGGradientRelease(grad);
    CGColorSpaceRelease(cs);
    CGContextRestoreGState(ctx);
}

static CTFontRef createFont(BOOL isMono, CGFloat size, CGFloat weight) {
    CFStringRef fontName;
    if (isMono) {
        fontName = (weight >= 0.5) ? CFSTR("Menlo-Bold") : CFSTR("Menlo-Regular");
    } else {
        if (weight >= 0.72) {
            fontName = CFSTR("AvenirNext-Bold");
        } else if (weight >= 0.52) {
            fontName = CFSTR("AvenirNext-DemiBold");
        } else {
            fontName = CFSTR("AvenirNext-Medium");
        }
    }
    return CTFontCreateWithName(fontName, size, NULL);
}

static CGFloat drawTextLine(CGContextRef ctx, NSString *str, CGFloat x, CGFloat y, CGFloat size, CGFloat weight, BOOL isMono, RGBA color, int align, CGFloat tracking) {
    if (!str || str.length == 0 || color.a <= 0.001) return 0;
    CTFontRef font = createFont(isMono, size, weight);
    CGColorSpaceRef cs = CGColorSpaceCreateDeviceRGB();
    CGFloat comps[4] = { color.r, color.g, color.b, color.a };
    CGColorRef cgColor = CGColorCreate(cs, comps);

    NSDictionary *attrs = @{
        (id)kCTFontAttributeName: (__bridge id)font,
        (id)kCTForegroundColorAttributeName: (__bridge id)cgColor,
        (id)kCTKernAttributeName: @(tracking)
    };
    NSAttributedString *attrStr = [[NSAttributedString alloc] initWithString:str attributes:attrs];
    CTLineRef line = CTLineCreateWithAttributedString((__bridge CFAttributedStringRef)attrStr);

    CGFloat ascent = 0, descent = 0, leading = 0;
    double width = CTLineGetTypographicBounds(line, &ascent, &descent, &leading);

    CGFloat drawX = x;
    if (align == 1) drawX = x - width * 0.5;
    else if (align == 2) drawX = x - width;

    CGContextSaveGState(ctx);
    CGContextSetTextMatrix(ctx, CGAffineTransformMakeScale(1.0, -1.0));
    CGContextSetTextPosition(ctx, drawX, y + ascent);
    CTLineDraw(line, ctx);
    CGContextRestoreGState(ctx);

    CFRelease(line);
    CGColorRelease(cgColor);
    CGColorSpaceRelease(cs);
    CFRelease(font);
    return (CGFloat)width;
}

static CGFloat measureTextLine(NSString *str, CGFloat size, CGFloat weight, BOOL isMono, CGFloat tracking) {
    if (!str || str.length == 0) return 0;
    CTFontRef font = createFont(isMono, size, weight);
    NSDictionary *attrs = @{
        (id)kCTFontAttributeName: (__bridge id)font,
        (id)kCTKernAttributeName: @(tracking)
    };
    NSAttributedString *attrStr = [[NSAttributedString alloc] initWithString:str attributes:attrs];
    CTLineRef line = CTLineCreateWithAttributedString((__bridge CFAttributedStringRef)attrStr);
    double width = CTLineGetTypographicBounds(line, NULL, NULL, NULL);
    CFRelease(line);
    CFRelease(font);
    return (CGFloat)width;
}

typedef enum {
    STATE_IDLE = 0,
    STATE_WORKING,
    STATE_THINKING,
    STATE_SEARCHING,
    STATE_APPROVAL,
    STATE_QUESTION,
    STATE_ERROR,
    STATE_FINISHED,
    STATE_RATELIMIT,
    STATE_SLEEPING
} MascotState;

typedef enum {
    MARK_NONE = 0,
    MARK_BAR,
    MARK_TRI,
    MARK_RING,
    MARK_DIAMOND,
    MARK_WAVE,
    MARK_DOT
} MascotMark;

typedef struct {
    const char *key;
    const char *label;
    uint32_t orbHex;
    uint32_t glowHex;
    int eyeType;
} StateVisual;

static StateVisual getStateVisual(MascotState st) {
    switch (st) {
        case STATE_IDLE:      return (StateVisual){ "idle", "Parado", 0x7b8499, 0xcfe3ff, 0 };
        case STATE_WORKING:   return (StateVisual){ "working", "Trabalhando", 0x4f8ef7, 0xbfe0ff, 0 };
        case STATE_THINKING:  return (StateVisual){ "thinking", "Pensando", 0x9b7bff, 0xe4d9ff, 0 };
        case STATE_SEARCHING: return (StateVisual){ "searching", "Buscando", 0x38c6e8, 0xc9f4ff, 0 };
        case STATE_APPROVAL:  return (StateVisual){ "approval", "Aguarda você", 0xf5b04a, 0xffe7b8, 1 };
        case STATE_QUESTION:  return (StateVisual){ "question", "Pergunta", 0x4fd1c5, 0xc8fff6, 0 };
        case STATE_ERROR:     return (StateVisual){ "error", "Bloqueado", 0xf2677a, 0xffc9d1, 3 };
        case STATE_FINISHED:  return (StateVisual){ "finished", "Concluído", 0x3ecf8e, 0xd1ffe6, 2 };
        case STATE_RATELIMIT: return (StateVisual){ "ratelimit", "Limite", 0xff9b4a, 0xffd9b8, 4 };
        case STATE_SLEEPING:  return (StateVisual){ "sleeping", "Dormindo", 0x4b5266, 0x8f9ab5, 4 };
    }
    return (StateVisual){ "idle", "Parado", 0x7b8499, 0xcfe3ff, 0 };
}

static void drawSparkleStar(CGContextRef ctx, CGFloat cx, CGFloat cy, CGFloat scale, RGBA color) {
    if (scale <= 0.01 || color.a <= 0.01) return;
    CGContextSaveGState(ctx);
    CGContextTranslateCTM(ctx, cx, cy);
    CGContextScaleCTM(ctx, scale, scale);
    CGContextBeginPath(ctx);
    CGContextMoveToPoint(ctx, 0, -4.5);
    CGContextAddLineToPoint(ctx, 1.2, -1.2);
    CGContextAddLineToPoint(ctx, 4.5, 0);
    CGContextAddLineToPoint(ctx, 1.2, 1.2);
    CGContextAddLineToPoint(ctx, 0, 4.5);
    CGContextAddLineToPoint(ctx, -1.2, 1.2);
    CGContextAddLineToPoint(ctx, -4.5, 0);
    CGContextAddLineToPoint(ctx, -1.2, -1.2);
    CGContextClosePath(ctx);
    setFill(ctx, color);
    CGContextFillPath(ctx);
    CGContextRestoreGState(ctx);
}

static void drawMascot(
    CGContextRef ctx,
    CGFloat x,
    CGFloat y,
    CGFloat size,
    uint32_t agentHex,
    MascotMark mark,
    MascotState state,
    CGFloat t,
    CGFloat phase,
    CGFloat extraLookX,
    CGFloat extraLookY,
    BOOL isPat,
    BOOL isWalking,
    CGFloat alpha
) {
    if (alpha <= 0.001) return;
    StateVisual sv = getStateVisual(state);
    RGBA agentCol = hexColor(agentHex, alpha);
    RGBA armCol = hexColor(agentHex, alpha);
    armCol.r *= 0.88; armCol.g *= 0.88; armCol.b *= 0.88;
    RGBA orbCol = hexColor(sv.orbHex, alpha);
    RGBA glowCol = hexColor(sv.glowHex, alpha);

    CGFloat scale = size / 120.0;
    CGFloat tp = t + phase;

    CGContextSaveGState(ctx);
    CGContextTranslateCTM(ctx, x, y);
    CGContextScaleCTM(ctx, scale, scale);

    CGFloat flX = 0.0;
    CGFloat flY = 0.0;
    CGFloat flRot = 0.0;
    CGFloat shadowScale = 1.0;

    if (isWalking) {
        flY = -4.0 * fabs(sin(tp * 7.0));
        flRot = 4.5 * sin(tp * 7.0) * (M_PI / 180.0);
        shadowScale = 1.0 - 0.12 * fabs(sin(tp * 7.0));
    } else if (isPat || state == STATE_FINISHED) {
        CGFloat cyc = fmod(tp * 1.35, 1.0);
        CGFloat hop = sin(cyc * M_PI);
        flY = -11.0 * hop;
        shadowScale = 1.0 - 0.18 * hop;
    } else if (state == STATE_APPROVAL) {
        CGFloat cyc = fmod(tp * 1.15, 1.0);
        CGFloat hop = pow(sin(cyc * M_PI), 1.6);
        flY = -12.0 * hop;
        shadowScale = 1.0 - 0.2 * hop;
    } else if (state == STATE_ERROR) {
        flX = 4.0 * sin(tp * 28.0) * exp(-fmod(tp, 1.2) * 2.5);
        flY = -2.0 * sin(tp * 2.4);
    } else {
        flY = -3.2 * sin(tp * 2.4);
        shadowScale = 1.0 - 0.08 * (0.5 + 0.5 * sin(tp * 2.4));
    }

    // 1. Ground shadow
    CGContextSaveGState(ctx);
    CGContextTranslateCTM(ctx, 60.0, 127.0);
    CGContextScaleCTM(ctx, shadowScale, 1.0);
    setFill(ctx, (RGBA){0, 0, 0, 0.38 * alpha});
    CGContextFillEllipseInRect(ctx, CGRectMake(-32.0, -4.5, 64.0, 9.0));
    CGContextRestoreGState(ctx);

    // Floating group
    CGContextSaveGState(ctx);
    CGContextTranslateCTM(ctx, 60.0 + flX, 75.0 + flY);
    CGContextRotateCTM(ctx, flRot);
    CGContextTranslateCTM(ctx, -60.0, -75.0);

    // 2. Arms
    CGFloat armWaveL = 0.0, armWaveR = 0.0;
    if (isPat || state == STATE_FINISHED) {
        armWaveL = 34.0 * (0.5 + 0.5 * sin(tp * 12.0)) * (M_PI / 180.0);
        armWaveR = -34.0 * (0.5 + 0.5 * sin(tp * 12.0)) * (M_PI / 180.0);
    } else if (state == STATE_APPROVAL) {
        armWaveL = 18.0 * sin(tp * 8.0) * (M_PI / 180.0);
        armWaveR = -18.0 * sin(tp * 8.0) * (M_PI / 180.0);
    }

    CGContextSaveGState(ctx);
    CGContextTranslateCTM(ctx, 13.0, 81.0);
    CGContextRotateCTM(ctx, armWaveL);
    CGContextTranslateCTM(ctx, -13.0, -81.0);
    setFill(ctx, armCol);
    CGContextFillEllipseInRect(ctx, CGRectMake(13.0 - 7.0, 90.0 - 11.0, 14.0, 22.0));
    CGContextRestoreGState(ctx);

    CGContextSaveGState(ctx);
    CGContextTranslateCTM(ctx, 107.0, 81.0);
    CGContextRotateCTM(ctx, armWaveR);
    CGContextTranslateCTM(ctx, -107.0, -81.0);
    setFill(ctx, armCol);
    CGContextFillEllipseInRect(ctx, CGRectMake(107.0 - 7.0, 90.0 - 11.0, 14.0, 22.0));
    CGContextRestoreGState(ctx);

    // 3. Antenna
    CGFloat antRot = 0.0;
    if (state == STATE_WORKING || state == STATE_SEARCHING) {
        antRot = 7.5 * sin(tp * 4.5) * (M_PI / 180.0);
    } else if (state == STATE_APPROVAL || state == STATE_QUESTION) {
        antRot = 10.0 * sin(tp * 9.0) * (M_PI / 180.0);
    }
    CGContextSaveGState(ctx);
    CGContextTranslateCTM(ctx, 60.0, 30.0);
    CGContextRotateCTM(ctx, antRot);
    CGContextTranslateCTM(ctx, -60.0, -30.0);

    setStroke(ctx, agentCol, 3.2);
    CGContextSetLineCap(ctx, kCGLineCapRound);
    CGContextBeginPath(ctx);
    CGContextMoveToPoint(ctx, 60.0, 30.0);
    CGContextAddLineToPoint(ctx, 60.0, 17.0);
    CGContextStrokePath(ctx);

    CGFloat orbPulse = 1.0;
    if (state == STATE_WORKING || state == STATE_THINKING || state == STATE_SEARCHING) {
        orbPulse = 0.75 + 0.25 * sin(tp * 4.2);
    } else if (state == STATE_APPROVAL) {
        orbPulse = 0.8 + 0.3 * fabs(sin(tp * 6.0));
    }
    RGBA glowOrb = orbCol;
    glowOrb.a = 0.45 * orbPulse * alpha;
    drawRadialGlow(ctx, 60.0, 11.0, 18.0, glowOrb);

    setFill(ctx, orbCol);
    CGContextFillEllipseInRect(ctx, CGRectMake(60.0 - 7.8, 11.0 - 7.8, 15.6, 15.6));

    RGBA darkGlyph = hexColor(0x10131a, alpha);
    if (state == STATE_WORKING || state == STATE_THINKING || state == STATE_SEARCHING) {
        for (int d = 0; d < 3; d++) {
            CGFloat dotPhase = fmod(tp * 3.0 - d * 0.25, 1.0);
            CGFloat dotAlpha = (0.35 + 0.65 * (0.5 + 0.5 * sin(dotPhase * 2 * M_PI))) * alpha;
            setFill(ctx, (RGBA){0.06, 0.07, 0.10, dotAlpha});
            CGFloat dx = 56.0 + d * 4.0;
            CGContextFillEllipseInRect(ctx, CGRectMake(dx - 1.4, 11.0 - 1.4, 2.8, 2.8));
        }
    } else if (state == STATE_APPROVAL) {
        drawTextLine(ctx, @"!", 60.0, 4.5, 11.5, 0.7, NO, darkGlyph, 1, 0);
    } else if (state == STATE_FINISHED) {
        setStroke(ctx, darkGlyph, 2.2);
        CGContextSetLineCap(ctx, kCGLineCapRound);
        CGContextSetLineJoin(ctx, kCGLineJoinRound);
        CGContextBeginPath(ctx);
        CGContextMoveToPoint(ctx, 56.4, 11.2);
        CGContextAddLineToPoint(ctx, 59.0, 13.8);
        CGContextAddLineToPoint(ctx, 63.6, 8.4);
        CGContextStrokePath(ctx);
    } else if (state == STATE_ERROR) {
        setStroke(ctx, darkGlyph, 2.2);
        CGContextSetLineCap(ctx, kCGLineCapRound);
        CGContextBeginPath(ctx);
        CGContextMoveToPoint(ctx, 57.2, 8.2);
        CGContextAddLineToPoint(ctx, 62.8, 13.8);
        CGContextMoveToPoint(ctx, 62.8, 8.2);
        CGContextAddLineToPoint(ctx, 57.2, 13.8);
        CGContextStrokePath(ctx);
    }
    CGContextRestoreGState(ctx);

    // 4. Body
    CGContextBeginPath(ctx);
    CGContextMoveToPoint(ctx, 60.0, 28.0);
    CGContextAddCurveToPoint(ctx, 86.0, 28.0, 102.0, 52.0, 105.0, 80.0);
    CGContextAddCurveToPoint(ctx, 108.0, 106.0, 92.0, 118.0, 60.0, 118.0);
    CGContextAddCurveToPoint(ctx, 28.0, 118.0, 12.0, 106.0, 15.0, 80.0);
    CGContextAddCurveToPoint(ctx, 18.0, 52.0, 34.0, 28.0, 60.0, 28.0);
    CGContextClosePath(ctx);
    setFill(ctx, agentCol);
    CGContextFillPath(ctx);

    // 5. Bottom shade
    CGContextBeginPath(ctx);
    CGContextMoveToPoint(ctx, 15.0, 80.0);
    CGContextAddCurveToPoint(ctx, 12.0, 106.0, 28.0, 118.0, 60.0, 118.0);
    CGContextAddCurveToPoint(ctx, 92.0, 118.0, 108.0, 106.0, 105.0, 80.0);
    CGContextAddCurveToPoint(ctx, 97.0, 98.0, 80.0, 106.0, 60.0, 106.0);
    CGContextAddCurveToPoint(ctx, 40.0, 106.0, 23.0, 98.0, 15.0, 80.0);
    CGContextClosePath(ctx);
    setFill(ctx, (RGBA){0, 0, 0, 0.21 * alpha});
    CGContextFillPath(ctx);

    // 6. Specular highlight
    CGContextSaveGState(ctx);
    CGContextTranslateCTM(ctx, 41.0, 50.0);
    CGContextRotateCTM(ctx, -30.0 * M_PI / 180.0);
    setFill(ctx, (RGBA){1, 1, 1, 0.36 * alpha});
    CGContextFillEllipseInRect(ctx, CGRectMake(-12.0, -6.0, 24.0, 12.0));
    CGContextRestoreGState(ctx);

    // 7. Chest mark
    RGBA mkCol = (RGBA){ 18.0/255.0, 24.0/255.0, 48.0/255.0, 0.55 * alpha };
    if (mark == MARK_BAR) {
        fillRoundRect(ctx, CGRectMake(52.0, 49.7, 16.0, 4.6), 2.3, mkCol);
    } else if (mark == MARK_TRI) {
        CGContextBeginPath(ctx);
        CGContextMoveToPoint(ctx, 60.0, 46.5);
        CGContextAddLineToPoint(ctx, 66.0, 57.0);
        CGContextAddLineToPoint(ctx, 54.0, 57.0);
        CGContextClosePath(ctx);
        setFill(ctx, mkCol);
        CGContextFillPath(ctx);
    } else if (mark == MARK_RING) {
        setStroke(ctx, mkCol, 2.6);
        CGContextStrokeEllipseInRect(ctx, CGRectMake(60.0 - 4.6, 52.0 - 4.6, 9.2, 9.2));
    } else if (mark == MARK_DIAMOND) {
        CGContextBeginPath(ctx);
        CGContextMoveToPoint(ctx, 60.0, 45.5);
        CGContextAddLineToPoint(ctx, 66.5, 52.0);
        CGContextAddLineToPoint(ctx, 60.0, 58.5);
        CGContextAddLineToPoint(ctx, 53.5, 52.0);
        CGContextClosePath(ctx);
        setFill(ctx, mkCol);
        CGContextFillPath(ctx);
    } else if (mark == MARK_WAVE) {
        setStroke(ctx, mkCol, 2.6);
        CGContextSetLineCap(ctx, kCGLineCapRound);
        CGContextBeginPath(ctx);
        CGContextMoveToPoint(ctx, 51.0, 53.0);
        CGContextAddQuadCurveToPoint(ctx, 55.5, 47.0, 60.0, 53.0);
        CGContextAddQuadCurveToPoint(ctx, 64.5, 59.0, 69.0, 53.0);
        CGContextStrokePath(ctx);
    }

    // 8. Visor
    fillAndStrokeRoundRect(ctx, CGRectMake(25.0, 70.0, 70.0, 32.0), 16.0,
                           hexColor(0x10131a, alpha),
                           (RGBA){1, 1, 1, 0.1 * alpha}, 1.0);

    // 9. Eyes
    CGFloat lookX = extraLookX;
    CGFloat lookY = extraLookY;
    CGFloat eyeRot = 0.0;
    if (state == STATE_WORKING) {
        lookX += 4.2 * sin(tp * 2.3);
    } else if (state == STATE_SEARCHING) {
        lookX += 5.5 * sin(tp * 4.8);
    } else if (state == STATE_THINKING) {
        lookX += 3.0;
        lookY -= 3.5;
    }

    CGFloat blinkCyc = fmod(tp, 4.2);
    CGFloat blinkScaleY = 1.0;
    if (blinkCyc > 3.95 && blinkCyc < 4.12) {
        CGFloat b = (blinkCyc - 3.95) / 0.17;
        blinkScaleY = 1.0 - 0.9 * sin(b * M_PI);
    }

    int eyeType = isPat ? 2 : sv.eyeType;

    CGContextSaveGState(ctx);
    CGContextTranslateCTM(ctx, 60.0 + lookX, 86.0 + lookY);
    CGContextRotateCTM(ctx, eyeRot);
    CGContextScaleCTM(ctx, 1.0, blinkScaleY);
    CGContextTranslateCTM(ctx, -60.0, -86.0);

    if (eyeType == 0) {
        fillRoundRect(ctx, CGRectMake(41.5, 78.0, 7.0, 16.0), 3.5, glowCol);
        fillRoundRect(ctx, CGRectMake(71.5, 78.0, 7.0, 16.0), 3.5, glowCol);
    } else if (eyeType == 1) {
        setFill(ctx, glowCol);
        CGContextFillEllipseInRect(ctx, CGRectMake(45.0 - 6.5, 86.0 - 6.5, 13.0, 13.0));
        CGContextFillEllipseInRect(ctx, CGRectMake(75.0 - 6.5, 86.0 - 6.5, 13.0, 13.0));
    } else if (eyeType == 2) {
        setStroke(ctx, glowCol, 4.0);
        CGContextSetLineCap(ctx, kCGLineCapRound);
        CGContextBeginPath(ctx);
        CGContextMoveToPoint(ctx, 39.0, 90.0);
        CGContextAddQuadCurveToPoint(ctx, 45.0, 79.0, 51.0, 90.0);
        CGContextMoveToPoint(ctx, 69.0, 90.0);
        CGContextAddQuadCurveToPoint(ctx, 75.0, 79.0, 81.0, 90.0);
        CGContextStrokePath(ctx);
    } else if (eyeType == 3) {
        setStroke(ctx, glowCol, 3.6);
        CGContextSetLineCap(ctx, kCGLineCapRound);
        CGContextBeginPath(ctx);
        CGContextMoveToPoint(ctx, 41.5, 82.0); CGContextAddLineToPoint(ctx, 48.5, 90.0);
        CGContextMoveToPoint(ctx, 48.5, 82.0); CGContextAddLineToPoint(ctx, 41.5, 90.0);
        CGContextMoveToPoint(ctx, 71.5, 82.0); CGContextAddLineToPoint(ctx, 78.5, 90.0);
        CGContextMoveToPoint(ctx, 78.5, 82.0); CGContextAddLineToPoint(ctx, 71.5, 90.0);
        CGContextStrokePath(ctx);
    }
    CGContextRestoreGState(ctx);
    CGContextRestoreGState(ctx);

    // 10. Sparkles on finished
    if (isPat || state == STATE_FINISHED) {
        CGFloat s1 = 0.35 + 0.95 * (0.5 + 0.5 * sin(tp * 5.5));
        CGFloat s2 = 0.35 + 0.85 * (0.5 + 0.5 * sin(tp * 5.5 - 1.2));
        CGFloat s3 = 0.35 + 0.75 * (0.5 + 0.5 * sin(tp * 5.5 - 2.4));
        drawSparkleStar(ctx, 100.0, 38.0, s1 * 1.15, (RGBA){1, 1, 1, alpha});
        drawSparkleStar(ctx, 20.0, 44.0, s2 * 0.95, (RGBA){1, 1, 1, alpha});
        drawSparkleStar(ctx, 106.0, 66.0, s3 * 0.85, (RGBA){1, 1, 1, alpha});
    }

    CGContextRestoreGState(ctx);
}

static CGFloat drawStatusChip(CGContextRef ctx, CGFloat x, CGFloat y, NSString *label, uint32_t dotHex, CGFloat alpha) {
    RGBA dotCol = hexColor(dotHex, alpha);
    RGBA bgCol = hexColor(dotHex, 0.14 * alpha);
    RGBA borderCol = hexColor(dotHex, 0.38 * alpha);
    RGBA textCol = mixColor(dotCol, hexColor(0xffffff, alpha), 0.55);
    textCol.a = alpha;

    CGFloat textW = measureTextLine(label, 14.0, 0.55, NO, 0);
    CGFloat pillW = 14.0 + 8.0 + 8.0 + textW + 15.0;
    CGFloat pillH = 30.0;

    fillAndStrokeRoundRect(ctx, CGRectMake(x, y, pillW, pillH), 15.0, bgCol, borderCol, 1.2);
    setFill(ctx, dotCol);
    CGContextFillEllipseInRect(ctx, CGRectMake(x + 12.0, y + (pillH - 8.0) * 0.5, 8.0, 8.0));
    drawTextLine(ctx, label, x + 27.0, y + 6.5, 14.0, 0.55, NO, textCol, 0, 0);
    return pillW;
}

static void drawCursor(CGContextRef ctx, CGFloat x, CGFloat y, BOOL clicking, CGFloat clickProgress, CGFloat alpha) {
    if (alpha <= 0.001) return;
    CGContextSaveGState(ctx);
    if (clicking && clickProgress > 0.0 && clickProgress < 1.0) {
        CGFloat r = 12.0 + clickProgress * 38.0;
        RGBA ringCol = hexColor(0x3ecf8e, (1.0 - clickProgress) * 0.80 * alpha);
        setStroke(ctx, ringCol, 3.0);
        CGContextStrokeEllipseInRect(ctx, CGRectMake(x - r, y - r, r * 2.0, r * 2.0));
    }
    CGContextTranslateCTM(ctx, x, y);
    if (clicking && clickProgress < 0.35) {
        CGContextScaleCTM(ctx, 0.88, 0.88);
    }
    CGContextBeginPath(ctx);
    CGContextMoveToPoint(ctx, 0, 0);
    CGContextAddLineToPoint(ctx, 0, 26);
    CGContextAddLineToPoint(ctx, 6.8, 20.2);
    CGContextAddLineToPoint(ctx, 11.5, 30.0);
    CGContextAddLineToPoint(ctx, 15.8, 28.0);
    CGContextAddLineToPoint(ctx, 11.0, 18.2);
    CGContextAddLineToPoint(ctx, 19.5, 18.2);
    CGContextClosePath(ctx);
    setFill(ctx, hexColor(0xffffff, alpha));
    CGContextFillPath(ctx);

    CGContextBeginPath(ctx);
    CGContextMoveToPoint(ctx, 0, 0);
    CGContextAddLineToPoint(ctx, 0, 26);
    CGContextAddLineToPoint(ctx, 6.8, 20.2);
    CGContextAddLineToPoint(ctx, 11.5, 30.0);
    CGContextAddLineToPoint(ctx, 15.8, 28.0);
    CGContextAddLineToPoint(ctx, 11.0, 18.2);
    CGContextAddLineToPoint(ctx, 19.5, 18.2);
    CGContextClosePath(ctx);
    setStroke(ctx, hexColor(0x0d0f14, alpha), 2.2);
    CGContextStrokePath(ctx);
    CGContextRestoreGState(ctx);
}

static void drawBackground(CGContextRef ctx, CGFloat t) {
    setFill(ctx, hexColor(0x0d0f14, 1.0));
    CGContextFillRect(ctx, CGRectMake(0, 0, WIDTH, HEIGHT));

    uint32_t glowHex = 0x4f8ef7;
    if (t >= 4.0 && t < 6.2) glowHex = 0xf5b04a;
    else if (t >= 6.2 && t < 9.0) glowHex = 0x3ecf8e;
    else if (t >= 14.5 && t < 17.2) glowHex = 0xf2677a;
    else if (t >= 17.2) glowHex = 0x3ecf8e;

    drawRadialGlow(ctx, WIDTH * 0.5, 140.0, 820.0, hexColor(glowHex, 0.11));
    drawRadialGlow(ctx, WIDTH * 0.5, HEIGHT * 0.58, 700.0, hexColor(0x4f8ef7, 0.05));

    setStroke(ctx, hexColor(0x4f8ef7, 0.05), 1.0);
    CGContextBeginPath(ctx);
    for (CGFloat gx = (WIDTH / 2) % 44; gx < WIDTH; gx += 44.0) {
        CGContextMoveToPoint(ctx, gx, 0);
        CGContextAddLineToPoint(ctx, gx, HEIGHT);
    }
    for (CGFloat gy = 0; gy < HEIGHT; gy += 44.0) {
        CGContextMoveToPoint(ctx, 0, gy);
        CGContextAddLineToPoint(ctx, WIDTH, gy);
    }
    CGContextStrokePath(ctx);
}

// ============================================================================
// SCENE 1 (0.0s - 4.0s): O PROBLEMA vs. AS 5 ESTAÇÕES ESPECIALIZADAS
// ============================================================================
static void renderScene1(CGContextRef ctx, CGFloat t) {
    CGFloat alpha = 0, offY = 0, sc = 1;
    sceneTransition(t, 0.0, 4.0, &alpha, &offY, &sc);
    if (alpha <= 0.001) return;

    CGContextSaveGState(ctx);
    CGContextTranslateCTM(ctx, WIDTH * 0.5, HEIGHT * 0.5 + offY);
    CGContextScaleCTM(ctx, sc, sc);
    CGContextTranslateCTM(ctx, -WIDTH * 0.5, -HEIGHT * 0.5);

    // Top Header
    drawTextLine(ctx, @"FORGE-SDD 3.0 · ARQUITETURA DE 5 ESTAÇÕES EM ESTEIRA",
                 WIDTH * 0.5, 42.0, 15.0, 0.65, YES, hexColor(0x4f8ef7, alpha), 1, 2.8);
    drawTextLine(ctx, @"De 1 agente sobrecarregado para 5 especialistas em esteira.",
                 WIDTH * 0.5, 76.0, 46.0, 0.8, NO, hexColor(0xe9ecf4, alpha), 1, -0.6);
    drawTextLine(ctx, @"Cada estação tem seu próprio papel, contexto limpo e contrato de entrega.",
                 WIDTH * 0.5, 138.0, 21.0, 0.45, NO, hexColor(0x99a2b8, alpha), 1, 0);

    // Problem vs Solution Banner (y = 188..288)
    CGRect bannerRect = CGRectMake(80.0, 188.0, 1760.0, 96.0);
    fillAndStrokeRoundRect(ctx, bannerRect, 20.0,
                           hexColor(0x12151d, 0.94 * alpha),
                           hexColor(0xf2677a, 0.30 * alpha), 1.4);
    drawMascot(ctx, 106.0, 194.0, 74.0, 0x7b8499, MARK_NONE, STATE_ERROR, t, 0.0, 0, 0, NO, NO, alpha);
    drawStatusChip(ctx, 198.0, 206.0, @"Antes: 1 Agente Fazendo Tudo", 0xf2677a, alpha);
    drawTextLine(ctx, @"Mistura spec com código · Esquece regras no meio da sessão · Aprova o próprio erro",
                 200.0, 246.0, 17.5, 0.55, NO, hexColor(0xd3d9e8, alpha), 0, 0);

    // Right badge inside banner
    fillAndStrokeRoundRect(ctx, CGRectMake(1355.0, 208.0, 455.0, 56.0), 16.0,
                           hexColor(0x3ecf8e, 0.14 * alpha),
                           hexColor(0x3ecf8e, 0.42 * alpha), 1.3);
    drawTextLine(ctx, @"AGORA: 5 ESTAÇÕES ISOLADAS ↓",
                 1582.5, 225.0, 16.5, 0.7, YES, hexColor(0x3ecf8e, alpha), 1, 0.8);

    // 5 Station Specialist Cards (y = 314..996)
    uint32_t cols[5] = { 0xf2a65a, 0x8aa4ff, 0xc9a7ff, 0x4fd1b5, 0xff8a5c };
    MascotMark marks[5] = { MARK_RING, MARK_DIAMOND, MARK_BAR, MARK_WAVE, MARK_TRI };
    MascotState states[5] = { STATE_WORKING, STATE_APPROVAL, STATE_THINKING, STATE_WORKING, STATE_SEARCHING };
    NSString *codes[5] = { @"ESTAÇÃO 01", @"ESTAÇÃO 02", @"ESTAÇÃO 03", @"ESTAÇÃO 04", @"ESTAÇÃO 05" };
    NSString *names[5] = { @"O Líder", @"PM & Arquiteto", @"O Planejador", @"Builders", @"Governança" };
    NSString *roles[5] = { @"Orquestrador Único", @"Spec & Reuse Check", @"Fatiador P / M / G", @"Execução em Fila", @"Auditor Read-Only" };
    NSString *desc1[5] = { @"Recebe seu pedido e", @"Vasculha o código atual", @"Quebra a spec aprovada", @"Consomem a fila em", @"Audita tudo sem poder" };
    NSString *desc2[5] = { @"coordena todo o time.", @"e pede seu OK na Spec.", @"em subtarefas na fila.", @"paralelo com testes.", @"editar 1 linha de código." };
    NSString *cmds[5]  = { @"/status · forge report", @"/discovery [ideia]", @"/split-features", @"/proxima-feature", @"/revisar · forge gate" };

    CGFloat cardW = 332.0;
    CGFloat cardH = 668.0;
    CGFloat gap = 25.0;
    CGFloat startX = 80.0;
    CGFloat cardY = 314.0;

    for (int i = 0; i < 5; i++) {
        CGFloat popStart = 0.45 + i * 0.22;
        CGFloat popU = easeOutBack(clampf((t - popStart) / 0.42, 0.0, 1.0));
        if (t < popStart) popU = 0.15;
        CGFloat cAlpha = alpha * clampf((t - (popStart - 0.15)) / 0.30, 0.25, 1.0);

        CGFloat cx = startX + i * (cardW + gap);
        CGFloat cy = cardY + (1.0 - popU) * 18.0;

        fillAndStrokeRoundRect(ctx, CGRectMake(cx, cy, cardW, cardH), 22.0,
                               hexColor(0x12151d, 0.96 * cAlpha),
                               hexColor(cols[i], 0.42 * cAlpha), 1.5);

        // Top color accent bar
        fillRoundRect(ctx, CGRectMake(cx + 22.0, cy + 18.0, 88.0, 5.0), 2.5, hexColor(cols[i], cAlpha));
        drawTextLine(ctx, codes[i], cx + 22.0, cy + 34.0, 13.5, 0.65, YES, hexColor(cols[i], cAlpha), 0, 1.4);

        StateVisual sv = getStateVisual(states[i]);
        drawStatusChip(ctx, cx + cardW - 158.0, cy + 26.0,
                       [NSString stringWithUTF8String:sv.label], sv.orbHex, cAlpha);

        // Mascot with radial glow
        drawRadialGlow(ctx, cx + cardW * 0.5, cy + 168.0, 115.0, hexColor(cols[i], 0.20 * cAlpha));
        drawMascot(ctx, cx + (cardW - 142.0) * 0.5, cy + 88.0, 142.0,
                   cols[i], marks[i], states[i], t, i * 0.6, 0, 0, NO, NO, cAlpha);

        // Station Name & Role
        drawTextLine(ctx, names[i], cx + cardW * 0.5, cy + 266.0, 27.0, 0.8, NO, hexColor(0xe9ecf4, cAlpha), 1, -0.3);

        // Role pill
        CGFloat rW = measureTextLine(roles[i], 14.5, 0.6, NO, 0) + 28.0;
        fillAndStrokeRoundRect(ctx, CGRectMake(cx + (cardW - rW) * 0.5, cy + 312.0, rW, 32.0), 16.0,
                               hexColor(cols[i], 0.14 * cAlpha),
                               hexColor(cols[i], 0.35 * cAlpha), 1.1);
        drawTextLine(ctx, roles[i], cx + cardW * 0.5, cy + 319.0, 14.5, 0.6, NO, hexColor(cols[i], cAlpha), 1, 0);

        // Educational description lines
        drawTextLine(ctx, desc1[i], cx + cardW * 0.5, cy + 372.0, 17.5, 0.45, NO, hexColor(0xd3d9e8, cAlpha), 1, 0);
        drawTextLine(ctx, desc2[i], cx + cardW * 0.5, cy + 400.0, 17.5, 0.45, NO, hexColor(0xd3d9e8, cAlpha), 1, 0);

        // Divider
        fillRoundRect(ctx, CGRectMake(cx + 24.0, cy + 452.0, cardW - 48.0, 1.0), 0.5, hexColor(0xffffff, 0.08 * cAlpha));

        // Responsibility contract mini-label
        drawTextLine(ctx, @"COMANDO PRINCIPAL", cx + cardW * 0.5, cy + 474.0, 12.0, 0.6, YES, hexColor(0x667089, cAlpha), 1, 1.5);

        // Command box at bottom
        fillAndStrokeRoundRect(ctx, CGRectMake(cx + 20.0, cy + 504.0, cardW - 40.0, 48.0), 12.0,
                               hexColor(0x0d0f14, 0.95 * cAlpha),
                               hexColor(0xffffff, 0.10 * cAlpha), 1.1);
        drawTextLine(ctx, cmds[i], cx + cardW * 0.5, cy + 518.0, 15.0, 0.6, YES, hexColor(0xe9ecf4, cAlpha), 1, 0);

        // Isolation badge at bottom of card
        NSString *iso = (i == 4) ? @"0 arquivos c/ escrita" : @"Contexto 100% isolado";
        uint32_t isoHex = (i == 4) ? 0xff8a5c : 0x3ecf8e;
        drawTextLine(ctx, iso, cx + cardW * 0.5, cy + 576.0, 14.0, 0.6, YES, hexColor(isoHex, cAlpha), 1, 0);
    }

    CGContextRestoreGState(ctx);
}

// ============================================================================
// SCENE 2 (4.0s - 9.0s): VISÃO PROCESSUAL & PORTÃO HUMANO (01 -> 02 -> 03 -> 04 -> 05 ↺)
// ============================================================================
static void renderScene2(CGContextRef ctx, CGFloat t) {
    CGFloat alpha = 0, offY = 0, sc = 1;
    sceneTransition(t, 4.0, 9.0, &alpha, &offY, &sc);
    if (alpha <= 0.001) return;

    CGContextSaveGState(ctx);
    CGContextTranslateCTM(ctx, WIDTH * 0.5, HEIGHT * 0.5 + offY);
    CGContextScaleCTM(ctx, sc, sc);
    CGContextTranslateCTM(ctx, -WIDTH * 0.5, -HEIGHT * 0.5);

    BOOL isApproved = (t >= 6.20);

    // Header
    drawTextLine(ctx, @"VISÃO PROCESSUAL · PASSAGEM DE BASTÃO ENTRE AS 5 ESTAÇÕES",
                 WIDTH * 0.5, 40.0, 15.0, 0.65, YES, hexColor(0x8aa4ff, alpha), 1, 2.5);
    drawTextLine(ctx, @"Você aprova a Spec no Portão FLX-01. A esteira executa o resto.",
                 WIDTH * 0.5, 72.0, 44.0, 0.8, NO, hexColor(0xe9ecf4, alpha), 1, -0.5);
    drawTextLine(ctx, @"Veja como cada estação recebe um insumo validado e entrega um artefato exato para a próxima:",
                 WIDTH * 0.5, 132.0, 20.5, 0.45, NO, hexColor(0x99a2b8, alpha), 1, 0);

    // 1. Connected Pipeline Stepper Card (y = 180..420)
    fillAndStrokeRoundRect(ctx, CGRectMake(80.0, 180.0, 1760.0, 236.0), 22.0,
                           hexColor(0x12151d, 0.96 * alpha),
                           hexColor(0xffffff, 0.11 * alpha), 1.4);

    uint32_t cols[5] = { 0xf2a65a, 0x8aa4ff, 0xc9a7ff, 0x4fd1b5, 0xff8a5c };
    MascotMark marks[5] = { MARK_RING, MARK_DIAMOND, MARK_BAR, MARK_WAVE, MARK_TRI };
    NSString *stNames[5] = { @"01 · O Líder", @"02 · PM & Arquiteto", @"03 · O Planejador", @"04 · Builders", @"05 · Governança" };
    NSString *stArtifacts[5] = { @"Intenção + Status", @"Spec + Portão FLX-01", @"Fila queue.json", @"Código + Testes", @"Gate + Read-Only" };

    CGFloat nodeW = 272.0;
    CGFloat nodeH = 142.0;
    CGFloat connW = 65.0;
    CGFloat startX = 120.0;
    CGFloat nodeY = 202.0;

    for (int i = 0; i < 5; i++) {
        CGFloat nx = startX + i * (nodeW + connW);

        MascotState st = STATE_IDLE;
        BOOL isFocus = NO;
        if (!isApproved) {
            if (i == 0) st = STATE_FINISHED;
            else if (i == 1) { st = STATE_APPROVAL; isFocus = YES; }
            else st = STATE_IDLE;
        } else {
            if (i == 0 || i == 1) st = STATE_FINISHED;
            else if (i == 2) { st = (t < 7.1) ? STATE_WORKING : STATE_FINISHED; isFocus = (t < 7.1); }
            else if (i == 3) { st = (t >= 7.1) ? STATE_WORKING : STATE_IDLE; isFocus = (t >= 7.1); }
            else st = STATE_SEARCHING;
        }

        uint32_t borderHex = isFocus ? cols[i] : (st == STATE_FINISHED ? 0x3ecf8e : 0xffffff);
        CGFloat borderAlpha = isFocus ? 0.85 : (st == STATE_FINISHED ? 0.55 : 0.12);

        fillAndStrokeRoundRect(ctx, CGRectMake(nx, nodeY, nodeW, nodeH), 18.0,
                               hexColor(isFocus ? 0x181c26 : 0x0d0f14, 0.95 * alpha),
                               hexColor(borderHex, borderAlpha * alpha), isFocus ? 2.2 : 1.2);

        drawMascot(ctx, nx + 10.0, nodeY + 14.0, 84.0, cols[i], marks[i], st, t, i * 0.4, 0, 0, NO, NO, alpha);

        drawTextLine(ctx, stNames[i], nx + 102.0, nodeY + 24.0, 15.0, 0.75, NO, hexColor(0xe9ecf4, alpha), 0, 0);
        drawTextLine(ctx, stArtifacts[i], nx + 102.0, nodeY + 50.0, 12.2, 0.5, YES, hexColor(cols[i], alpha), 0, 0);

        StateVisual sv = getStateVisual(st);
        drawStatusChip(ctx, nx + 100.0, nodeY + 84.0, [NSString stringWithUTF8String:sv.label], sv.orbHex, alpha);

        // Draw connector arrow to next station
        if (i < 4) {
            CGFloat ax1 = nx + nodeW + 8.0;
            CGFloat ax2 = nx + nodeW + connW - 8.0;
            CGFloat ay = nodeY + nodeH * 0.5;

            BOOL linkActive = (!isApproved && i == 0) || (isApproved && i <= 2);
            RGBA lineCol = hexColor(linkActive ? 0x3ecf8e : 0x4f8ef7, (linkActive ? 0.75 : 0.28) * alpha);
            setStroke(ctx, lineCol, 3.0);
            CGContextBeginPath(ctx);
            CGContextMoveToPoint(ctx, ax1, ay);
            CGContextAddLineToPoint(ctx, ax2, ay);
            CGContextStrokePath(ctx);

            // Arrowhead
            CGContextBeginPath(ctx);
            CGContextMoveToPoint(ctx, ax2, ay);
            CGContextAddLineToPoint(ctx, ax2 - 9.0, ay - 6.0);
            CGContextAddLineToPoint(ctx, ax2 - 9.0, ay + 6.0);
            CGContextClosePath(ctx);
            setFill(ctx, lineCol);
            CGContextFillPath(ctx);

            // Animated data packet dot along connector
            CGFloat pktPhase = fmod(t * 1.6 - i * 0.3, 1.0);
            if (pktPhase < 0) pktPhase += 1.0;
            CGFloat px = ax1 + pktPhase * (ax2 - ax1);
            setFill(ctx, hexColor(linkActive ? 0x3ecf8e : cols[i], alpha));
            CGContextFillEllipseInRect(ctx, CGRectMake(px - 4.5, ay - 4.5, 9.0, 9.0));
        }
    }

    // Bottom return loop pill inside pipeline card
    fillAndStrokeRoundRect(ctx, CGRectMake(490.0, 358.0, 940.0, 40.0), 20.0,
                           hexColor(0xff8a5c, 0.12 * alpha),
                           hexColor(0xff8a5c, 0.38 * alpha), 1.2);
    drawTextLine(ctx, @"LOOP DE AUTOCORREÇÃO (05 -> 04): Se a Governança reprovar, devolve direto para a fila do Builder corrigir",
                 WIDTH * 0.5, 369.0, 14.0, 0.6, YES, hexColor(0xff8a5c, alpha), 1, 0);

    // 2. Bottom 3-Column Procedural Contract Cards (y = 444..992)
    CGFloat colW = 564.0;
    CGFloat colH = 544.0;
    CGFloat colGap = 34.0;
    CGFloat colY = 444.0;

    // COLUMN 1: RECEBE (ENTRADA)
    CGFloat c1X = 80.0;
    fillAndStrokeRoundRect(ctx, CGRectMake(c1X, colY, colW, colH), 22.0,
                           hexColor(0x12151d, 0.96 * alpha),
                           hexColor(0x8aa4ff, 0.35 * alpha), 1.4);
    drawStatusChip(ctx, c1X + 28.0, colY + 26.0, @"1 · O QUE RECEBE (ENTRADA)", 0x8aa4ff, alpha);
    drawTextLine(ctx, @"Insumo Limpo + Reuse Check", c1X + 28.0, colY + 74.0, 26.0, 0.8, NO, hexColor(0xe9ecf4, alpha), 0, -0.3);
    drawTextLine(ctx, @"Antes de propor qualquer arquivo novo, a Estação 02", c1X + 28.0, colY + 118.0, 17.0, 0.45, NO, hexColor(0x99a2b8, alpha), 0, 0);
    drawTextLine(ctx, @"vasculha o repositório para reaproveitar o que já existe:", c1X + 28.0, colY + 144.0, 17.0, 0.45, NO, hexColor(0x99a2b8, alpha), 0, 0);

    // Terminal box inside Column 1
    fillAndStrokeRoundRect(ctx, CGRectMake(c1X + 28.0, colY + 186.0, colW - 56.0, 326.0), 16.0,
                           hexColor(0x0d0f14, 0.96 * alpha),
                           hexColor(0xffffff, 0.09 * alpha), 1.1);
    drawTextLine(ctx, @"$ /discovery módulo de fila determinística", c1X + 48.0, colY + 210.0, 15.0, 0.65, YES, hexColor(0x8aa4ff, alpha), 0, 0);
    drawTextLine(ctx, @"[01 -> 02] O Líder acionou PM & Arquiteto", c1X + 48.0, colY + 248.0, 14.0, 0.45, YES, hexColor(0x99a2b8, alpha), 0, 0);
    drawTextLine(ctx, @"✓ Reuse Check (DUP-04): cmd/forge-sdd/main.go", c1X + 48.0, colY + 284.0, 14.0, 0.55, YES, hexColor(0x3ecf8e, alpha), 0, 0);
    drawTextLine(ctx, @"✓ Constitution verificada: ARQ-01..07, FLX-01", c1X + 48.0, colY + 316.0, 14.0, 0.55, YES, hexColor(0x3ecf8e, alpha), 0, 0);
    drawTextLine(ctx, @"✓ Gerados: discovery.md, criteria.md, plan.md", c1X + 48.0, colY + 348.0, 14.0, 0.55, YES, hexColor(0x3ecf8e, alpha), 0, 0);
    fillRoundRect(ctx, CGRectMake(c1X + 48.0, colY + 392.0, colW - 96.0, 96.0), 12.0, hexColor(0x181c26, alpha));
    drawTextLine(ctx, @"GARANTIA DE ENTRADA:", c1X + 66.0, colY + 410.0, 13.0, 0.7, YES, hexColor(0x8aa4ff, alpha), 0, 1.0);
    drawTextLine(ctx, @"Zero código é escrito sem antes mapear o projeto", c1X + 66.0, colY + 436.0, 15.0, 0.5, NO, hexColor(0xe9ecf4, alpha), 0, 0);
    drawTextLine(ctx, @"e alinhar os critérios de aceite com você.", c1X + 66.0, colY + 460.0, 15.0, 0.5, NO, hexColor(0xe9ecf4, alpha), 0, 0);

    // COLUMN 2: PORTÃO HUMANO OBRIGATÓRIO (FLX-01)
    CGFloat c2X = c1X + colW + colGap;
    uint32_t c2Hex = isApproved ? 0x3ecf8e : 0xf5b04a;
    fillAndStrokeRoundRect(ctx, CGRectMake(c2X, colY, colW, colH), 22.0,
                           hexColor(0x12151d, 0.96 * alpha),
                           hexColor(c2Hex, 0.55 * alpha), 1.8);
    drawStatusChip(ctx, c2X + 28.0, colY + 26.0,
                   isApproved ? @"2 · PORTÃO HUMANO FLX-01 (APROVADO)" : @"2 · PORTÃO HUMANO FLX-01 (AGUARDA VOCÊ)",
                   c2Hex, alpha);
    drawTextLine(ctx, @"Seu 'De Acordo' na Especificação", c2X + 28.0, colY + 74.0, 26.0, 0.8, NO, hexColor(0xe9ecf4, alpha), 0, -0.3);
    drawTextLine(ctx, @"A esteira trava obrigatoriamente na Estação 02 até", c2X + 28.0, colY + 118.0, 17.0, 0.45, NO, hexColor(0x99a2b8, alpha), 0, 0);
    drawTextLine(ctx, @"você validar a arquitetura e os critérios Gherkin:", c2X + 28.0, colY + 144.0, 17.0, 0.45, NO, hexColor(0x99a2b8, alpha), 0, 0);

    // Mascot 02 + Spec preview card inside Column 2
    fillAndStrokeRoundRect(ctx, CGRectMake(c2X + 28.0, colY + 186.0, colW - 56.0, 212.0), 16.0,
                           hexColor(0x0d0f14, 0.96 * alpha),
                           hexColor(c2Hex, 0.28 * alpha), 1.2);
    drawMascot(ctx, c2X + 46.0, colY + 206.0, 118.0, 0x8aa4ff, MARK_DIAMOND,
               isApproved ? STATE_FINISHED : STATE_APPROVAL, t, 0.2, 0, 0, isApproved, NO, alpha);
    drawTextLine(ctx, @"Spec pronta para revisão:", c2X + 182.0, colY + 214.0, 17.0, 0.7, NO, hexColor(0xe9ecf4, alpha), 0, 0);
    drawTextLine(ctx, @"• sdd/discovery/discovery-58.md", c2X + 182.0, colY + 248.0, 14.5, 0.55, YES, hexColor(0xd3d9e8, alpha), 0, 0);
    drawTextLine(ctx, @"• sdd/discovery/criteria-58.md", c2X + 182.0, colY + 278.0, 14.5, 0.55, YES, hexColor(0xd3d9e8, alpha), 0, 0);
    drawTextLine(ctx, @"• sdd/discovery/plan-58.md", c2X + 182.0, colY + 308.0, 14.5, 0.55, YES, hexColor(0xd3d9e8, alpha), 0, 0);
    drawTextLine(ctx, isApproved ? @"Status: De Acordo confirmado pelo usuário!" : @"Status: Aguardando seu 'De Acordo' (FLX-01)",
                 c2X + 182.0, colY + 348.0, 14.5, 0.65, NO, hexColor(c2Hex, alpha), 0, 0);

    // Interactive Approval Button at bottom of Column 2
    CGRect btnRect = CGRectMake(c2X + 28.0, colY + 422.0, colW - 56.0, 86.0);
    fillAndStrokeRoundRect(ctx, btnRect, 16.0,
                           hexColor(c2Hex, isApproved ? 0.22 * alpha : 0.16 * alpha),
                           hexColor(c2Hex, 0.75 * alpha), 1.8);
    drawTextLine(ctx,
                 isApproved ? @"✓ SPEC APROVADA · BASTÃO PASSADO PARA ESTAÇÃO 03" : @"▶ CLICAR PARA APROVAR SPEC (PORTÃO FLX-01)",
                 c2X + colW * 0.5, colY + 452.0, 16.5, 0.75, YES, hexColor(0xe9ecf4, alpha), 1, 0.4);

    // COLUMN 3: ENTREGA (SAÍDA & FILA)
    CGFloat c3X = c2X + colW + colGap;
    uint32_t c3Hex = isApproved ? 0x4fd1b5 : 0xc9a7ff;
    fillAndStrokeRoundRect(ctx, CGRectMake(c3X, colY, colW, colH), 22.0,
                           hexColor(0x12151d, 0.96 * alpha),
                           hexColor(c3Hex, 0.38 * alpha), 1.4);
    drawStatusChip(ctx, c3X + 28.0, colY + 26.0, @"3 · O QUE ENTREGA (SAÍDA)", c3Hex, alpha);
    drawTextLine(ctx, @"Fila Fatiada P/M/G + Builders", c3X + 28.0, colY + 74.0, 26.0, 0.8, NO, hexColor(0xe9ecf4, alpha), 0, -0.3);
    drawTextLine(ctx, @"Com a Spec aprovada, o Planejador (03) fatia o plano", c3X + 28.0, colY + 118.0, 17.0, 0.45, NO, hexColor(0x99a2b8, alpha), 0, 0);
    drawTextLine(ctx, @"em subtarefas isoladas para os Builders (04):", c3X + 28.0, colY + 144.0, 17.0, 0.45, NO, hexColor(0x99a2b8, alpha), 0, 0);

    fillAndStrokeRoundRect(ctx, CGRectMake(c3X + 28.0, colY + 186.0, colW - 56.0, 326.0), 16.0,
                           hexColor(0x0d0f14, 0.96 * alpha),
                           hexColor(0xffffff, 0.09 * alpha), 1.1);
    drawTextLine(ctx, @"sdd/.queue/queue.json (Fila Determinística)", c3X + 48.0, colY + 210.0, 15.0, 0.65, YES, hexColor(0xc9a7ff, alpha), 0, 0);

    NSString *qItems[3] = {
        @"feat-01 [P] Constitution & Rules   -> COMPLETED",
        @"feat-02 [M] Papéis .agents/roles/  -> BUILDER #1",
        @"feat-03 [M] CLI Go queue & gate    -> BUILDER #2"
    };
    uint32_t qCols[3] = { 0x3ecf8e, 0x4fd1b5, 0x4f8ef7 };
    for (int q = 0; q < 3; q++) {
        CGFloat qy = colY + 252.0 + q * 54.0;
        fillAndStrokeRoundRect(ctx, CGRectMake(c3X + 46.0, qy, colW - 92.0, 42.0), 10.0,
                               hexColor(qCols[q], isApproved ? 0.14 * alpha : 0.06 * alpha),
                               hexColor(qCols[q], isApproved ? 0.40 * alpha : 0.18 * alpha), 1.0);
        drawTextLine(ctx, qItems[q], c3X + 62.0, qy + 12.0, 14.0, 0.6, YES, hexColor(0xe9ecf4, alpha), 0, 0);
    }

    drawTextLine(ctx, @"✓ Cada tarefa já nasce com suas .agents/rules/ vinculadas",
                 c3X + 48.0, colY + 430.0, 14.5, 0.55, NO, hexColor(0x3ecf8e, alpha), 0, 0);
    drawTextLine(ctx, @"✓ Sem arquivos sobrepostos = paralelismo seguro",
                 c3X + 48.0, colY + 460.0, 14.5, 0.55, NO, hexColor(0x4fd1b5, alpha), 0, 0);

    // Animated Cursor moving to click "Aprovar Spec (Portão FLX-01)" at t = 6.20s
    if (t >= 4.8 && t <= 7.2) {
        CGFloat moveU = smoothstep(4.8, 6.05, t);
        CGFloat curX = 640.0 + moveU * (c2X + colW * 0.5 - 640.0);
        CGFloat curY = 360.0 + moveU * (colY + 465.0 - 360.0);
        BOOL isClicking = (t >= 6.15 && t <= 6.65);
        CGFloat clickProg = clampf((t - 6.15) / 0.45, 0.0, 1.0);
        drawCursor(ctx, curX, curY, isClicking, clickProg, alpha);
    }

    CGContextRestoreGState(ctx);
}

// ============================================================================
// SCENE 3 (9.0s - 14.5s): COMO INSTALAR E USAR NA PRÁTICA (4 COMANDOS REAIS)
// ============================================================================
static void renderScene3(CGContextRef ctx, CGFloat t) {
    CGFloat alpha = 0, offY = 0, sc = 1;
    sceneTransition(t, 9.0, 14.5, &alpha, &offY, &sc);
    if (alpha <= 0.001) return;

    CGContextSaveGState(ctx);
    CGContextTranslateCTM(ctx, WIDTH * 0.5, HEIGHT * 0.5 + offY);
    CGContextScaleCTM(ctx, sc, sc);
    CGContextTranslateCTM(ctx, -WIDTH * 0.5, -HEIGHT * 0.5);

    // Header
    drawTextLine(ctx, @"GUIA PRÁTICO · COMO INSTALAR E UTILIZAR NO SEU REPOSITÓRIO",
                 WIDTH * 0.5, 38.0, 15.0, 0.65, YES, hexColor(0x4fd1b5, alpha), 1, 2.5);
    drawTextLine(ctx, @"Instale via NPM. Conduza o time com 4 comandos simples.",
                 WIDTH * 0.5, 70.0, 44.0, 0.8, NO, hexColor(0xe9ecf4, alpha), 1, -0.5);

    // Top Installation Bar (y = 142..242)
    fillAndStrokeRoundRect(ctx, CGRectMake(80.0, 142.0, 1760.0, 96.0), 20.0,
                           hexColor(0x12151d, 0.96 * alpha),
                           hexColor(0x4f8ef7, 0.38 * alpha), 1.5);
    drawStatusChip(ctx, 110.0, 160.0, @"PASSO 0 · INSTALAÇÃO RÁPIDA", 0x3ecf8e, alpha);
    drawTextLine(ctx, @"Global ou Direto via NPX:", 112.0, 200.0, 16.5, 0.6, NO, hexColor(0x99a2b8, alpha), 0, 0);

    // Command pill 1
    fillAndStrokeRoundRect(ctx, CGRectMake(380.0, 158.0, 650.0, 64.0), 14.0,
                           hexColor(0x0d0f14, 0.98 * alpha),
                           hexColor(0xffffff, 0.12 * alpha), 1.1);
    drawTextLine(ctx, @"$ npm install -g @nathanramorim/forge-sdd && forge init",
                 404.0, 180.0, 16.5, 0.65, YES, hexColor(0x3ecf8e, alpha), 0, 0);

    // Command pill 2
    fillAndStrokeRoundRect(ctx, CGRectMake(1055.0, 158.0, 525.0, 64.0), 14.0,
                           hexColor(0x0d0f14, 0.98 * alpha),
                           hexColor(0xffffff, 0.12 * alpha), 1.1);
    drawTextLine(ctx, @"$ npx @nathanramorim/forge-sdd@latest init",
                 1079.0, 180.0, 16.5, 0.65, YES, hexColor(0x8aa4ff, alpha), 0, 0);

    drawTextLine(ctx, @"MIT · Open Source", 1705.0, 180.0, 15.5, 0.7, YES, hexColor(0xf2a65a, alpha), 1, 0);

    // Determine active step (0..3) based on time (9.0..14.5s)
    int activeStep = 0;
    if (t >= 13.0) activeStep = 3;
    else if (t >= 11.8) activeStep = 2;
    else if (t >= 10.5) activeStep = 1;
    else activeStep = 0;

    // LEFT COLUMN (x = 80..980): Os 4 Comandos do Dia a Dia no Chat da IDE
    CGFloat leftX = 80.0;
    CGFloat leftW = 900.0;
    CGFloat panelY = 264.0;
    CGFloat panelH = 732.0;

    fillAndStrokeRoundRect(ctx, CGRectMake(leftX, panelY, leftW, panelH), 22.0,
                           hexColor(0x12151d, 0.96 * alpha),
                           hexColor(0xffffff, 0.11 * alpha), 1.4);
    drawTextLine(ctx, @"NO CHAT DA SUA IDE (CLAUDE CODE, GEMINI CLI OU COPILOT)",
                 leftX + 32.0, panelY + 26.0, 14.0, 0.65, YES, hexColor(0x99a2b8, alpha), 0, 1.4);
    drawTextLine(ctx, @"Os 4 Comandos do Dia a Dia",
                 leftX + 32.0, panelY + 54.0, 26.0, 0.8, NO, hexColor(0xe9ecf4, alpha), 0, -0.3);

    NSString *stepCmds[4] = {
        @"1. /discovery [sua ideia]",
        @"2. /split-features",
        @"3. /proxima-feature",
        @"4. /revisar"
    };
    NSString *stepStations[4] = {
        @"Estação 02 · PM & Arquiteto",
        @"Estação 03 · O Planejador",
        @"Estação 04 · Builders em Fila",
        @"Estação 05 · Governança Read-Only"
    };
    NSString *stepDescs[4] = {
        @"Faz o Reuse Check do código atual, cria Spec + Critérios e pede seu OK no Portão FLX-01.",
        @"Quebra o plano técnico aprovado em subtarefas P/M/G e popula a fila sdd/.queue/queue.json.",
        @"Aciona os Builders para implementarem código + testes (em paralelo quando sem conflito).",
        @"Roda o Portão Determinístico (forge-sdd gate) + Auditoria Somente-Leitura e entrega o resumo."
    };
    uint32_t stepColors[4] = { 0x8aa4ff, 0xc9a7ff, 0x4fd1b5, 0xff8a5c };

    for (int s = 0; s < 4; s++) {
        CGFloat sy = panelY + 108.0 + s * 148.0;
        BOOL isCurrent = (s == activeStep);
        BOOL isDone = (s < activeStep);

        uint32_t bCol = isCurrent ? stepColors[s] : (isDone ? 0x3ecf8e : 0xffffff);
        CGFloat bAlpha = isCurrent ? 0.85 : (isDone ? 0.45 : 0.10);

        fillAndStrokeRoundRect(ctx, CGRectMake(leftX + 28.0, sy, leftW - 56.0, 132.0), 16.0,
                               hexColor(isCurrent ? 0x181c26 : 0x0d0f14, 0.96 * alpha),
                               hexColor(bCol, bAlpha * alpha), isCurrent ? 2.2 : 1.2);

        drawTextLine(ctx, stepCmds[s], leftX + 52.0, sy + 20.0, 21.0, 0.75, YES,
                     hexColor(isCurrent ? stepColors[s] : 0xe9ecf4, alpha), 0, 0);

        drawStatusChip(ctx, leftX + leftW - 330.0, sy + 16.0,
                       stepStations[s], isDone ? 0x3ecf8e : stepColors[s], alpha);

        drawTextLine(ctx, stepDescs[s], leftX + 52.0, sy + 68.0, 15.5, 0.45, NO,
                     hexColor(isCurrent ? 0xe9ecf4 : 0x99a2b8, alpha), 0, 0);
    }

    // RIGHT COLUMN (x = 1010..1840): Emulador ao Vivo do Motor CLI Go + Mascote da Estação Ativa
    CGFloat rightX = 1010.0;
    CGFloat rightW = 830.0;

    fillAndStrokeRoundRect(ctx, CGRectMake(rightX, panelY, rightW, panelH), 22.0,
                           hexColor(0x12151d, 0.96 * alpha),
                           hexColor(stepColors[activeStep], 0.48 * alpha), 1.6);

    // Top Mascot Showcase inside Right Column
    MascotMark activeMarks[4] = { MARK_DIAMOND, MARK_BAR, MARK_WAVE, MARK_TRI };
    MascotState activeStates[4] = { STATE_APPROVAL, STATE_THINKING, STATE_WORKING, STATE_FINISHED };

    drawRadialGlow(ctx, rightX + 120.0, panelY + 125.0, 110.0, hexColor(stepColors[activeStep], 0.22 * alpha));
    drawMascot(ctx, rightX + 46.0, panelY + 42.0, 138.0,
               stepColors[activeStep], activeMarks[activeStep], activeStates[activeStep],
               t, 0.3, 0, 0, NO, NO, alpha);

    drawTextLine(ctx, @"O QUE ACONTECE NOS BASTIDORES (CLI GO)",
                 rightX + 210.0, panelY + 46.0, 13.5, 0.65, YES, hexColor(stepColors[activeStep], alpha), 0, 1.4);
    drawTextLine(ctx, stepStations[activeStep],
                 rightX + 210.0, panelY + 74.0, 28.0, 0.8, NO, hexColor(0xe9ecf4, alpha), 0, -0.3);
    drawTextLine(ctx, @"Motor binário forge-sdd garante estado determinístico:",
                 rightX + 210.0, panelY + 116.0, 16.5, 0.45, NO, hexColor(0x99a2b8, alpha), 0, 0);

    // Live Terminal Window inside Right Column
    CGRect termRect = CGRectMake(rightX + 30.0, panelY + 206.0, rightW - 60.0, 496.0);
    fillAndStrokeRoundRect(ctx, termRect, 18.0,
                           hexColor(0x0d0f14, 0.98 * alpha),
                           hexColor(0xffffff, 0.12 * alpha), 1.2);

    // Terminal title bar dots
    setFill(ctx, hexColor(0xf2677a, alpha));
    CGContextFillEllipseInRect(ctx, CGRectMake(termRect.origin.x + 22.0, termRect.origin.y + 18.0, 11.0, 11.0));
    setFill(ctx, hexColor(0xf5b04a, alpha));
    CGContextFillEllipseInRect(ctx, CGRectMake(termRect.origin.x + 40.0, termRect.origin.y + 18.0, 11.0, 11.0));
    setFill(ctx, hexColor(0x3ecf8e, alpha));
    CGContextFillEllipseInRect(ctx, CGRectMake(termRect.origin.x + 58.0, termRect.origin.y + 18.0, 11.0, 11.0));
    drawTextLine(ctx, @"forge-sdd — queue · gate · skills",
                 termRect.origin.x + termRect.size.width * 0.5, termRect.origin.y + 15.0, 13.5, 0.55, YES, hexColor(0x667089, alpha), 1, 0);

    CGFloat tx = termRect.origin.x + 28.0;
    CGFloat ty = termRect.origin.y + 58.0;

    drawTextLine(ctx, @"$ forge-sdd queue list", tx, ty, 15.5, 0.7, YES, hexColor(0x4f8ef7, alpha), 0, 0);
    drawTextLine(ctx, @"ID          SIZE   STATUS       STATION / OWNER", tx, ty + 34.0, 14.0, 0.6, YES, hexColor(0x667089, alpha), 0, 0);
    drawTextLine(ctx, @"feat-58-01  [P]    COMPLETED    Builder #1 (Constitution & Rules)", tx, ty + 64.0, 14.5, 0.55, YES, hexColor(0x3ecf8e, alpha), 0, 0);
    drawTextLine(ctx, @"feat-58-02  [M]    COMPLETED    Builder #2 (Papéis .agents/roles/)", tx, ty + 94.0, 14.5, 0.55, YES, hexColor(0x3ecf8e, alpha), 0, 0);
    drawTextLine(ctx, @"feat-58-04  [M]    COMPLETED    Builder #3 (CLI Go queue & gate)", tx, ty + 124.0, 14.5, 0.55, YES, hexColor(0x3ecf8e, alpha), 0, 0);

    drawTextLine(ctx, @"$ forge-sdd gate --feature feat-58", tx, ty + 174.0, 15.5, 0.7, YES, hexColor(0x4f8ef7, alpha), 0, 0);
    drawTextLine(ctx, @"✓ [Branch Check]   feat/58-orquestracao-autonoma (OK)", tx, ty + 208.0, 14.5, 0.55, YES, hexColor(0x3ecf8e, alpha), 0, 0);
    drawTextLine(ctx, @"✓ [Spec Linked]    discovery-58 + criteria-58 + plan-58 (OK)", tx, ty + 238.0, 14.5, 0.55, YES, hexColor(0x3ecf8e, alpha), 0, 0);
    drawTextLine(ctx, @"✓ [Secret Scan]    0 chaves ou credenciais expostas (SEG-01)", tx, ty + 268.0, 14.5, 0.55, YES, hexColor(0x3ecf8e, alpha), 0, 0);
    drawTextLine(ctx, @"✓ [Test Suite]     go test ./... (100% PASS)", tx, ty + 298.0, 14.5, 0.55, YES, hexColor(0x3ecf8e, alpha), 0, 0);

    fillAndStrokeRoundRect(ctx, CGRectMake(tx, ty + 344.0, termRect.size.width - 56.0, 68.0), 12.0,
                           hexColor(0x3ecf8e, 0.14 * alpha),
                           hexColor(0x3ecf8e, 0.42 * alpha), 1.2);
    drawTextLine(ctx, @"RESULTADO: Paralelismo real sem colisão de arquivos e portão 100% determinístico.",
                 tx + 20.0, ty + 368.0, 15.0, 0.65, NO, hexColor(0xe9ecf4, alpha), 0, 0);

    CGContextRestoreGState(ctx);
}

// ============================================================================
// SCENE 4 (14.5s - 20.0s): GOVERNANÇA EM 3 CAMADAS & AUTO-SKILLS (READ-ONLY)
// ============================================================================
static void renderScene4(CGContextRef ctx, CGFloat t) {
    CGFloat alpha = 0, offY = 0, sc = 1;
    sceneTransition(t, 14.5, 20.0, &alpha, &offY, &sc);
    if (alpha <= 0.001) return;

    CGContextSaveGState(ctx);
    CGContextTranslateCTM(ctx, WIDTH * 0.5, HEIGHT * 0.5 + offY);
    CGContextScaleCTM(ctx, sc, sc);
    CGContextTranslateCTM(ctx, -WIDTH * 0.5, -HEIGHT * 0.5);

    BOOL isFixed = (t >= 17.25);

    // Header
    drawTextLine(ctx, @"GOVERNANÇA EM 3 CAMADAS · AUDITORIA SOMENTE-LEITURA + AUTO-SKILLS",
                 WIDTH * 0.5, 40.0, 15.0, 0.65, YES, hexColor(isFixed ? 0x3ecf8e : 0xff8a5c, alpha), 1, 2.5);
    drawTextLine(ctx, @"Quem escreve o código nunca aprova o próprio trabalho.",
                 WIDTH * 0.5, 72.0, 45.0, 0.8, NO, hexColor(0xe9ecf4, alpha), 1, -0.5);
    drawTextLine(ctx, @"A Estação 05 tem zero permissão de escrita: se encontrar falha, devolve para a fila do Builder (05 -> 04).",
                 WIDTH * 0.5, 132.0, 20.5, 0.45, NO, hexColor(0x99a2b8, alpha), 1, 0);

    // LEFT PANEL (x = 80..760): Simulador do Loop 05 -> 04
    CGFloat leftX = 80.0;
    CGFloat leftW = 680.0;
    CGFloat panelY = 188.0;
    CGFloat panelH = 804.0;

    uint32_t stateHex = isFixed ? 0x3ecf8e : 0xf2677a;
    fillAndStrokeRoundRect(ctx, CGRectMake(leftX, panelY, leftW, panelH), 22.0,
                           hexColor(0x12151d, 0.96 * alpha),
                           hexColor(stateHex, 0.48 * alpha), 1.6);

    drawStatusChip(ctx, leftX + 32.0, panelY + 28.0,
                   isFixed ? @"FASE 2 · CORRIGIDO PELO BUILDER & APROVADO" : @"FASE 1 · VIOLAÇÃO BLOQUEADA PELA ESTAÇÃO 05",
                   stateHex, alpha);

    drawTextLine(ctx, isFixed ? @"Loop 05 -> 04 Concluído com Sucesso" : @"Conflito de Interesse Eliminado na Raiz",
                 leftX + 32.0, panelY + 78.0, 26.0, 0.8, NO, hexColor(0xe9ecf4, alpha), 0, -0.3);

    // Two Mascots interacting inside Left Panel: Station 05 (Governance) & Station 04 (Builder)
    CGRect stageRect = CGRectMake(leftX + 32.0, panelY + 132.0, leftW - 64.0, 310.0);
    fillAndStrokeRoundRect(ctx, stageRect, 18.0,
                           hexColor(0x0d0f14, 0.96 * alpha),
                           hexColor(0xffffff, 0.10 * alpha), 1.2);

    // Station 05 Mascot (Left side of stage)
    drawRadialGlow(ctx, stageRect.origin.x + 145.0, stageRect.origin.y + 130.0, 110.0,
                   hexColor(isFixed ? 0x3ecf8e : 0xff8a5c, 0.22 * alpha));
    drawMascot(ctx, stageRect.origin.x + 75.0, stageRect.origin.y + 42.0, 140.0,
               0xff8a5c, MARK_TRI, isFixed ? STATE_FINISHED : STATE_ERROR,
               t, 0.1, 4.0, 0, isFixed, NO, alpha);
    drawTextLine(ctx, @"05 · Governança", stageRect.origin.x + 145.0, stageRect.origin.y + 212.0, 17.5, 0.75, NO, hexColor(0xe9ecf4, alpha), 1, 0);
    drawTextLine(ctx, @"0 arquivos c/ escrita", stageRect.origin.x + 145.0, stageRect.origin.y + 240.0, 13.5, 0.6, YES, hexColor(0xff8a5c, alpha), 1, 0);

    // Animated Loop Arrow between 05 and 04
    CGFloat midX = stageRect.origin.x + stageRect.size.width * 0.5;
    CGFloat midY = stageRect.origin.y + 125.0;
    fillAndStrokeRoundRect(ctx, CGRectMake(midX - 82.0, midY - 24.0, 164.0, 48.0), 24.0,
                           hexColor(stateHex, 0.16 * alpha),
                           hexColor(stateHex, 0.55 * alpha), 1.4);
    drawTextLine(ctx, isFixed ? @"04 -> 05 PASS" : @"05 -> 04 REPROVA",
                 midX, midY - 9.0, 14.5, 0.75, YES, hexColor(stateHex, alpha), 1, 0.5);

    // Station 04 Mascot (Right side of stage)
    drawRadialGlow(ctx, stageRect.origin.x + stageRect.size.width - 145.0, stageRect.origin.y + 130.0, 110.0,
                   hexColor(0x4fd1b5, 0.22 * alpha));
    drawMascot(ctx, stageRect.origin.x + stageRect.size.width - 215.0, stageRect.origin.y + 42.0, 140.0,
               0x4fd1b5, MARK_WAVE, isFixed ? STATE_FINISHED : STATE_WORKING,
               t, 0.6, -4.0, 0, isFixed, NO, alpha);
    drawTextLine(ctx, @"04 · Builder", stageRect.origin.x + stageRect.size.width - 145.0, stageRect.origin.y + 212.0, 17.5, 0.75, NO, hexColor(0xe9ecf4, alpha), 1, 0);
    drawTextLine(ctx, isFixed ? @"Correção aplicada!" : @"Corrigindo na fila...",
                 stageRect.origin.x + stageRect.size.width - 145.0, stageRect.origin.y + 240.0, 13.5, 0.6, YES, hexColor(0x4fd1b5, alpha), 1, 0);

    // Code Diff Preview inside Left Panel (y = panelY + 468)
    CGRect diffRect = CGRectMake(leftX + 32.0, panelY + 468.0, leftW - 64.0, 302.0);
    fillAndStrokeRoundRect(ctx, diffRect, 16.0,
                           hexColor(0x0d0f14, 0.98 * alpha),
                           hexColor(0xffffff, 0.10 * alpha), 1.1);
    drawTextLine(ctx, @"internal/config/auth.go — Inspecionado pela Governança",
                 diffRect.origin.x + 22.0, diffRect.origin.y + 18.0, 14.0, 0.6, YES, hexColor(0x99a2b8, alpha), 0, 0);

    // Red removed/blocked line
    fillRoundRect(ctx, CGRectMake(diffRect.origin.x + 18.0, diffRect.origin.y + 56.0, diffRect.size.width - 36.0, 44.0), 8.0,
                  hexColor(0xf2677a, isFixed ? 0.10 * alpha : 0.22 * alpha));
    drawTextLine(ctx, @"- const apiKey = \"sk-live-98a7f6d5c4e3b2a1\" // [SEG-01 BLOQUEADO]",
                 diffRect.origin.x + 30.0, diffRect.origin.y + 69.0, 14.0, 0.6, YES, hexColor(0xf2677a, alpha), 0, 0);

    // Green fixed line
    fillRoundRect(ctx, CGRectMake(diffRect.origin.x + 18.0, diffRect.origin.y + 112.0, diffRect.size.width - 36.0, 44.0), 8.0,
                  hexColor(0x3ecf8e, isFixed ? 0.22 * alpha : 0.06 * alpha));
    drawTextLine(ctx, @"+ apiKey := os.Getenv(\"FORGE_API_KEY\")      // [SEG-01 SEGURO]",
                 diffRect.origin.x + 30.0, diffRect.origin.y + 125.0, 14.0, 0.6, YES,
                 hexColor(isFixed ? 0x3ecf8e : 0x667089, alpha), 0, 0);

    drawTextLine(ctx, @"POR QUE ISSO IMPORTA?", diffRect.origin.x + 22.0, diffRect.origin.y + 184.0, 13.0, 0.7, YES, hexColor(0xff8a5c, alpha), 0, 1.2);
    drawTextLine(ctx, @"O Auditor roda em contexto Read-Only separado do Builder.", diffRect.origin.x + 22.0, diffRect.origin.y + 212.0, 16.0, 0.5, NO, hexColor(0xe9ecf4, alpha), 0, 0);
    drawTextLine(ctx, @"Ele nunca 'passa pano' para código duplicado ou sem teste.", diffRect.origin.x + 22.0, diffRect.origin.y + 240.0, 16.0, 0.5, NO, hexColor(0xd3d9e8, alpha), 0, 0);

    // RIGHT PANEL (x = 790..1840): As 3 Camadas de Governança + Auto-Skills
    CGFloat rightX = 790.0;
    CGFloat rightW = 1050.0;

    NSString *layerTitles[3] = {
        @"CAMADA A · PORTÃO DETERMINÍSTICO (CLI GO: forge-sdd gate)",
        @"CAMADA B · ANALISTA DE GOVERNANÇA (SUBAGENTE SOMENTE-LEITURA)",
        @"CAMADA C · PAINEL DO LÍDER & AUTO-SKILLS (forge-sdd skills)"
    };
    NSString *layerSub[3] = {
        isFixed ? @"PASS — Branch válida, Spec vinculada, 0 segredos expostos, testes Go 100% verdes."
                : @"FAIL [SEG-01 / P-13] — Chave hardcoded detectada no diff. Pull Request bloqueado.",
        isFixed ? @"APROVADO — Código 100% aderente à Constitution (ARQ/SEG/DUP) e às Rules de domínio."
                : @"REPROVADO — Item devolvido automaticamente para a fila do Builder (Loop 05 -> 04).",
        isFixed ? @"AUTO-SKILL CRIADA (DUP-05) — Procedimento salvo em .agents/skills/ e sincronizado!"
                : @"AGUARDANDO CAMADAS A & B — O Líder só apresenta o PR quando os portões estão verdes."
    };
    NSString *layerDetail[3] = {
        @"Verifica em milissegundos: git branch != main, arquivos da spec, regex de segredos, limite de linhas e suite de testes.",
        @"Compara cada requisito Gherkin (criteria.md) com a implementação real e audita regressões arquiteturais.",
        @"Padrões que se repetem 3+ vezes viram Skills reutilizáveis (.agents/, .claude/, .gemini/) com aprovação humana."
    };

    for (int l = 0; l < 3; l++) {
        CGFloat ly = panelY + l * 276.0;
        CGFloat lh = 252.0;

        uint32_t lHex = isFixed ? (l == 2 ? 0xf2a65a : 0x3ecf8e) : (l < 2 ? 0xf2677a : 0xf5b04a);
        fillAndStrokeRoundRect(ctx, CGRectMake(rightX, ly, rightW, lh), 22.0,
                               hexColor(0x12151d, 0.96 * alpha),
                               hexColor(lHex, 0.48 * alpha), 1.6);

        drawStatusChip(ctx, rightX + 32.0, ly + 26.0, layerTitles[l], lHex, alpha);

        drawTextLine(ctx, layerSub[l], rightX + 32.0, ly + 78.0, 22.0, 0.75, NO, hexColor(0xe9ecf4, alpha), 0, -0.2);
        drawTextLine(ctx, layerDetail[l], rightX + 32.0, ly + 122.0, 16.5, 0.45, NO, hexColor(0x99a2b8, alpha), 0, 0);

        // Bottom badge strip inside each Layer card
        fillAndStrokeRoundRect(ctx, CGRectMake(rightX + 32.0, ly + 168.0, rightW - 64.0, 56.0), 12.0,
                               hexColor(0x0d0f14, 0.95 * alpha),
                               hexColor(lHex, 0.28 * alpha), 1.1);
        NSString *cmdLine = (l == 0)
            ? (isFixed ? @"$ forge-sdd gate --feature feat-58   ->   Exit Code 0 (100% APROVADO)"
                       : @"$ forge-sdd gate --feature feat-58   ->   Exit Code 1 (VIOLAÇÃO SEG-01 BLOQUEADA)")
            : (l == 1)
                ? @"Permissão de Escrita: NENHUMA (Read-Only)   ·   Regras auditadas: ARQ, SEG, DUP, COD, TST"
                : @"$ forge-sdd skills sync   ->   Sincronizado: .agents/skills/ · .claude/ · .gemini/";
        drawTextLine(ctx, cmdLine, rightX + 52.0, ly + 187.0, 15.0, 0.6, YES, hexColor(lHex, alpha), 0, 0);
    }

    CGContextRestoreGState(ctx);
}

// ============================================================================
// SCENE 5 (20.0s - 24.0s): OUTRO — OS 5 ESPECIALISTAS + FORGE-SDD.ORG
// ============================================================================
static void renderScene5(CGContextRef ctx, CGFloat t) {
    CGFloat alpha = 0, offY = 0, sc = 1;
    sceneTransition(t, 20.0, 24.0, &alpha, &offY, &sc);
    if (alpha <= 0.001) return;

    CGContextSaveGState(ctx);
    CGContextTranslateCTM(ctx, WIDTH * 0.5, HEIGHT * 0.5 + offY);
    CGContextScaleCTM(ctx, sc, sc);
    CGContextTranslateCTM(ctx, -WIDTH * 0.5, -HEIGHT * 0.5);

    // Top status pill
    CGFloat pillW = 620.0;
    CGFloat pillX = (WIDTH - pillW) * 0.5;
    fillAndStrokeRoundRect(ctx, CGRectMake(pillX, 34.0, pillW, 44.0), 22.0,
                           hexColor(0x0c0e13, 0.95 * alpha),
                           hexColor(0x3ecf8e, 0.45 * alpha), 1.4);
    setFill(ctx, hexColor(0x3ecf8e, alpha));
    CGContextFillEllipseInRect(ctx, CGRectMake(pillX + 24.0, 51.0, 10.0, 10.0));
    drawTextLine(ctx, @"Forge-SDD 3.0 · 5 estações em esteira · Open Source (MIT)",
                 WIDTH * 0.5 + 10.0, 47.0, 15.0, 0.6, YES, hexColor(0xe9ecf4, alpha), 1, 0);

    // Center row of 5 celebrating Mascots
    uint32_t cols[5] = { 0xf2a65a, 0x8aa4ff, 0xc9a7ff, 0x4fd1b5, 0xff8a5c };
    MascotMark marks[5] = { MARK_RING, MARK_DIAMOND, MARK_BAR, MARK_WAVE, MARK_TRI };
    NSString *names[5] = { @"01 · O Líder", @"02 · PM & Arquiteto", @"03 · O Planejador", @"04 · Builders", @"05 · Governança" };

    CGFloat botSize = 154.0;
    CGFloat botGap = 62.0;
    CGFloat totalBotsW = 5 * botSize + 4 * botGap;
    CGFloat startBX = (WIDTH - totalBotsW) * 0.5;
    CGFloat botY = 126.0;

    for (int i = 0; i < 5; i++) {
        CGFloat bx = startBX + i * (botSize + botGap);
        drawRadialGlow(ctx, bx + botSize * 0.5, botY + botSize * 0.5, 120.0, hexColor(cols[i], 0.22 * alpha));
        drawMascot(ctx, bx, botY, botSize, cols[i], marks[i], STATE_FINISHED,
                   t, i * 0.45, 0, 0, YES, NO, alpha);
        drawTextLine(ctx, names[i], bx + botSize * 0.5, botY + 182.0, 17.5, 0.7, NO, hexColor(0xe9ecf4, alpha), 1, 0);
    }

    // Brand Headline & Punchline
    drawTextLine(ctx, @"FORGE-SDD 3.0 · HTTPS://FORGE-SDD.ORG",
                 WIDTH * 0.5, 378.0, 18.0, 0.7, YES, hexColor(0x4f8ef7, alpha), 1, 3.5);
    drawTextLine(ctx, @"Especificação clara. Execução em fila. Governança real.",
                 WIDTH * 0.5, 420.0, 56.0, 0.82, NO, hexColor(0xe9ecf4, alpha), 1, -0.8);
    drawTextLine(ctx, @"Transforme agentes de IA em uma esteira de engenharia confiável e auditável.",
                 WIDTH * 0.5, 504.0, 26.0, 0.55, NO, hexColor(0x3ecf8e, alpha), 1, -0.2);

    // Command Showcase Box (y = 578..768)
    CGFloat boxW = 1420.0;
    CGFloat boxX = (WIDTH - boxW) * 0.5;
    fillAndStrokeRoundRect(ctx, CGRectMake(boxX, 578.0, boxW, 190.0), 24.0,
                           hexColor(0x12151d, 0.96 * alpha),
                           hexColor(0x4f8ef7, 0.42 * alpha), 1.6);

    drawTextLine(ctx, @"COMECE AGORA NO SEU TERMINAL", boxX + 42.0, 606.0, 13.5, 0.7, YES, hexColor(0x8aa4ff, alpha), 0, 1.5);
    fillAndStrokeRoundRect(ctx, CGRectMake(boxX + 42.0, 640.0, 760.0, 96.0), 16.0,
                           hexColor(0x0d0f14, 0.98 * alpha),
                           hexColor(0x3ecf8e, 0.45 * alpha), 1.4);
    drawTextLine(ctx, @"$ npx @nathanramorim/forge-sdd@latest init",
                 boxX + 70.0, 662.0, 22.0, 0.7, YES, hexColor(0x3ecf8e, alpha), 0, 0);
    drawTextLine(ctx, @"# Depois no chat: /discovery -> /split-features -> /proxima-feature -> /revisar",
                 boxX + 70.0, 700.0, 14.0, 0.5, YES, hexColor(0x99a2b8, alpha), 0, 0);

    // Right side of box: URL & GitHub
    drawTextLine(ctx, @"SITE OFICIAL & DOCUMENTAÇÃO", boxX + 845.0, 606.0, 13.5, 0.7, YES, hexColor(0xf2a65a, alpha), 0, 1.5);
    drawTextLine(ctx, @"forge-sdd.org", boxX + 845.0, 642.0, 38.0, 0.82, NO, hexColor(0xe9ecf4, alpha), 0, -0.5);
    drawTextLine(ctx, @"github.com/nathanramorim/forge-sdd", boxX + 845.0, 696.0, 18.0, 0.6, YES, hexColor(0x8aa4ff, alpha), 0, 0);

    // Bottom Platform Chips Bar (y = 804..884)
    CGFloat barW = 1420.0;
    CGFloat barX = (WIDTH - barW) * 0.5;
    fillAndStrokeRoundRect(ctx, CGRectMake(barX, 804.0, barW, 80.0), 20.0,
                           hexColor(0x12151d, 0.92 * alpha),
                           hexColor(0xffffff, 0.10 * alpha), 1.2);

    CGFloat c1 = drawStatusChip(ctx, barX + 36.0, 829.0, @"Claude Code", 0xf2a65a, alpha);
    CGFloat c2 = drawStatusChip(ctx, barX + 36.0 + c1 + 16.0, 829.0, @"Gemini CLI & Antigravity", 0x4f8ef7, alpha);
    CGFloat c3 = drawStatusChip(ctx, barX + 36.0 + c1 + c2 + 32.0, 829.0, @"GitHub Copilot", 0xc9a7ff, alpha);
    drawStatusChip(ctx, barX + 36.0 + c1 + c2 + c3 + 48.0, 829.0, @"CLI Nativa em Go", 0x3ecf8e, alpha);

    drawTextLine(ctx, @"Open Source · Licença MIT · @nathanramorim/forge-sdd",
                 barX + barW - 36.0, 835.0, 15.5, 0.55, YES, hexColor(0xd3d9e8, alpha), 2, 0);

    CGContextRestoreGState(ctx);
}

static void renderFrameAtTime(CGContextRef ctx, CGFloat t) {
    CGContextSaveGState(ctx);
    CGContextTranslateCTM(ctx, 0, HEIGHT);
    CGContextScaleCTM(ctx, 1.0, -1.0);

    drawBackground(ctx, t);

    if (t < 4.0) {
        renderScene1(ctx, t);
    } else if (t < 9.0) {
        renderScene2(ctx, t);
    } else if (t < 14.5) {
        renderScene3(ctx, t);
    } else if (t < 20.0) {
        renderScene4(ctx, t);
    } else {
        renderScene5(ctx, t);
    }

    CGContextRestoreGState(ctx);
}

static void saveContextToFile(CGContextRef ctx, NSString *path, BOOL asJpeg) {
    CGImageRef img = CGBitmapContextCreateImage(ctx);
    NSURL *url = [NSURL fileURLWithPath:path];
    CFStringRef uti = asJpeg ? CFSTR("public.jpeg") : CFSTR("public.png");
    CGImageDestinationRef dest = CGImageDestinationCreateWithURL((__bridge CFURLRef)url, uti, 1, NULL);
    if (dest) {
        NSDictionary *props = asJpeg ? @{ (id)kCGImageDestinationLossyCompressionQuality: @(0.95) } : nil;
        CGImageDestinationAddImage(dest, img, (__bridge CFDictionaryRef)props);
        CGImageDestinationFinalize(dest);
        CFRelease(dest);
    }
    CGImageRelease(img);
}

int main(int argc, const char * argv[]) {
    @autoreleasepool {
        CGColorSpaceRef cs = CGColorSpaceCreateDeviceRGB();
        uint8_t *rawBuf = (uint8_t *)calloc(WIDTH * HEIGHT * 4, 1);
        CGContextRef ctx = CGBitmapContextCreate(
            rawBuf, WIDTH, HEIGHT, 8, WIDTH * 4, cs,
            kCGImageAlphaPremultipliedLast | kCGBitmapByteOrder32Big
        );
        CGContextSetAllowsAntialiasing(ctx, true);
        CGContextSetShouldAntialias(ctx, true);
        CGContextSetInterpolationQuality(ctx, kCGInterpolationHigh);

        if (argc > 1 && strcmp(argv[1], "--stills") == 0) {
            double checkTimes[] = { 2.20, 5.20, 7.40, 11.20, 13.50, 15.80, 18.40, 21.80 };
            int n = sizeof(checkTimes) / sizeof(checkTimes[0]);
            for (int i = 0; i < n; i++) {
                renderFrameAtTime(ctx, checkTimes[i]);
                NSString *p = [NSString stringWithFormat:@"/Users/nathanramorim/git/forge-sdd/brag-output/work/still_%02d_%.2fs.png", i, checkTimes[i]];
                saveContextToFile(ctx, p, NO);
                fprintf(stderr, "Saved still: %s\n", [p UTF8String]);
            }
            renderFrameAtTime(ctx, POSTER_TIME);
            saveContextToFile(ctx, @"/Users/nathanramorim/git/forge-sdd/brag-output/brag.jpg", YES);
            fprintf(stderr, "Saved poster: /Users/nathanramorim/git/forge-sdd/brag-output/brag.jpg\n");
        } else if (argc > 1 && strcmp(argv[1], "--stream") == 0) {
            for (int f = 0; f < TOTAL_FRAMES; f++) {
                CGFloat t = (f == 0) ? POSTER_TIME : ((CGFloat)f / (CGFloat)FPS);
                renderFrameAtTime(ctx, t);
                fwrite(rawBuf, 1, WIDTH * HEIGHT * 4, stdout);
                if (f % 90 == 0 || f == TOTAL_FRAMES - 1) {
                    fprintf(stderr, "Rendered frame %d / %d (t=%.2fs)\n", f + 1, TOTAL_FRAMES, t);
                }
            }
            fflush(stdout);
        }

        CGContextRelease(ctx);
        CGColorSpaceRelease(cs);
        free(rawBuf);
    }
    return 0;
}
