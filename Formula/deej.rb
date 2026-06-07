class Deej < Formula
  desc "Open-source hardware volume mixer for Windows and Linux"
  homepage "https://github.com/mxkolb30/homebrew-deej"
  url "https://github.com/mxkolb30/homebrew-deej/archive/fc3766584db783d52be10f9d3863da2c3dca214b.tar.gz"
  version "0.9.10"
  sha256 "89d293703787f5465e5a01bcbf682f9df095c2b74fe13066dc3c4c3504fb2075"
  license "MIT"

  head "https://github.com/mxkolb30/homebrew-deej.git", branch: "master"

  depends_on "go" => :build
  depends_on "pkg-config" => :build
  depends_on "gtk+3"
  depends_on "webkitgtk"
  depends_on "libayatana-appindicator"

  def install
    # Create compatibility symlinks for pkg-config
    (buildpath/"pkgconfig").mkdir
    
    # libayatana-appindicator provides ayatana-appindicator3-0.1.pc
    # but systray looks for appindicator3-0.1.pc
    ln_s Formula["libayatana-appindicator"].opt_lib/"pkgconfig/ayatana-appindicator3-0.1.pc", 
         buildpath/"pkgconfig/appindicator3-0.1.pc"
    
    # Create compatibility symlinks for headers
    (buildpath/"include").mkdir
    ln_s Formula["libayatana-appindicator"].opt_include/"libayatana-appindicator3-0.1/libayatana-appindicator",
         buildpath/"include/libappindicator"
    
    # Check for webkit2gtk version and create compatibility symlink if needed
    webkit_pc = Formula["webkitgtk"].opt_lib/"pkgconfig/webkit2gtk-4.1.pc"
    if webkit_pc.exist?
      ln_s webkit_pc, buildpath/"pkgconfig/webkit2gtk-4.0.pc"
    end
    
    ENV.append_path "PKG_CONFIG_PATH", buildpath/"pkgconfig"
    ENV.append "CGO_CFLAGS", "-I#{buildpath}/include"

    # Shove git commit, version tag into env if building from head
    commit = Utils.git_short_head if build.head?
    version_tag = version.to_s
    
    ldflags = [
      "-s -w",
      "-X main.gitCommit=#{commit}",
      "-X main.versionTag=#{version_tag}",
      "-X main.buildType=release"
    ].join(" ")

    system "go", "build", "-o", "deej", "-ldflags", ldflags, "./pkg/deej/cmd"
    
    bin.install "deej"
    
    # Install default config
    (etc/"deej").install "config.yaml"
  end

  def caveats
    <<~EOS
      deej requires a configuration file to map your sliders.
      A default configuration has been installed to:
        #{etc}/deej/config.yaml

      You should copy it to your home directory to customize it:
        mkdir -p ~/.config/deej
        cp #{etc}/deej/config.yaml ~/.config/deej/config.yaml

      To run deej, simply type `deej` in your terminal.
      If you are on Bazzite or another immutable OS, ensure you have
      the necessary tray icon support (e.g., AppIndicator extension).
    EOS
  end

  test do
    system "#{bin}/deej", "--version"
  end
end
