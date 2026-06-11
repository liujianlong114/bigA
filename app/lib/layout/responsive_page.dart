import 'package:flutter/material.dart';
import 'app_breakpoints.dart';

/// 页面级响应式容器：桌面居中限宽，手机全宽
class ResponsivePage extends StatelessWidget {
  final Widget child;
  final double? maxWidth;
  final EdgeInsetsGeometry? padding;
  final bool scrollable;

  const ResponsivePage({
    super.key,
    required this.child,
    this.maxWidth,
    this.padding,
    this.scrollable = true,
  });

  @override
  Widget build(BuildContext context) {
    final pad = padding ?? AppBreakpoints.pagePadding(context);
    final limit = maxWidth ?? (AppBreakpoints.isDesktop(context) ? AppBreakpoints.contentMax : double.infinity);

    Widget body = Align(
      alignment: Alignment.topCenter,
      child: ConstrainedBox(
        constraints: BoxConstraints(maxWidth: limit),
        child: child,
      ),
    );

    if (scrollable) {
      body = SingleChildScrollView(
        physics: const AlwaysScrollableScrollPhysics(parent: BouncingScrollPhysics()),
        padding: pad,
        child: body,
      );
    } else {
      body = Padding(padding: pad, child: body);
    }
    return body;
  }
}

/// 桌面左右分栏
class DesktopSplitView extends StatelessWidget {
  final Widget primary;
  final Widget secondary;
  final int primaryFlex;
  final int secondaryFlex;
  final double gap;
  final bool scrollSecondary;

  const DesktopSplitView({
    super.key,
    required this.primary,
    required this.secondary,
    this.primaryFlex = 4,
    this.secondaryFlex = 6,
    this.gap = 20,
    this.scrollSecondary = true,
  });

  @override
  Widget build(BuildContext context) {
    if (AppBreakpoints.isMobile(context)) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [primary, SizedBox(height: gap), secondary],
      );
    }

    final right = scrollSecondary ? SingleChildScrollView(child: secondary) : secondary;

    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Expanded(flex: primaryFlex, child: primary),
        SizedBox(width: gap),
        Expanded(flex: secondaryFlex, child: right),
      ],
    );
  }
}

/// 桌面多列网格
class ResponsiveGrid extends StatelessWidget {
  final List<Widget> children;
  final int desktopColumns;
  final int mobileColumns;
  final double gap;

  const ResponsiveGrid({
    super.key,
    required this.children,
    this.desktopColumns = 2,
    this.mobileColumns = 1,
    this.gap = 12,
  });

  @override
  Widget build(BuildContext context) {
    final cols = AppBreakpoints.gridColumns(context, desktop: desktopColumns, mobile: mobileColumns);
    if (cols <= 1) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [for (var i = 0; i < children.length; i++) ...[if (i > 0) SizedBox(height: gap), children[i]]],
      );
    }

    return LayoutGrid(columns: cols, gap: gap, children: children);
  }
}

class LayoutGrid extends StatelessWidget {
  final int columns;
  final double gap;
  final List<Widget> children;

  const LayoutGrid({
    super.key,
    required this.columns,
    required this.gap,
    required this.children,
  });

  @override
  Widget build(BuildContext context) {
    final rows = <Widget>[];
    for (var i = 0; i < children.length; i += columns) {
      if (rows.isNotEmpty) rows.add(SizedBox(height: gap));
      rows.add(
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            for (var j = 0; j < columns; j++) ...[
              if (j > 0) SizedBox(width: gap),
              Expanded(
                child: i + j < children.length ? children[i + j] : const SizedBox.shrink(),
              ),
            ],
          ],
        ),
      );
    }
    return Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: rows);
  }
}

/// 按钮组：桌面左对齐自适应宽度，手机可全宽
class AdaptiveButtonRow extends StatelessWidget {
  final List<Widget> buttons;
  final bool expandOnMobile;
  final double spacing;

  const AdaptiveButtonRow({
    super.key,
    required this.buttons,
    this.expandOnMobile = true,
    this.spacing = 10,
  });

  @override
  Widget build(BuildContext context) {
    final mobile = AppBreakpoints.isMobile(context);
    if (mobile && expandOnMobile && buttons.length == 1) {
      return SizedBox(width: double.infinity, child: buttons.first);
    }

    return Wrap(
      alignment: WrapAlignment.start,
      spacing: spacing,
      runSpacing: spacing,
      children: buttons,
    );
  }
}

/// 分段选择：桌面紧凑，手机均分
class AdaptiveSegmentRow extends StatelessWidget {
  final List<Widget> segments;

  const AdaptiveSegmentRow({super.key, required this.segments});

  @override
  Widget build(BuildContext context) {
    if (AppBreakpoints.isMobile(context)) {
      return Row(
        children: [
          for (var i = 0; i < segments.length; i++) ...[
            if (i > 0) const SizedBox(width: 8),
            Expanded(child: segments[i]),
          ],
        ],
      );
    }
    return Wrap(spacing: 8, runSpacing: 8, children: segments);
  }
}
